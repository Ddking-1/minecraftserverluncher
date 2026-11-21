// Account administration endpoints (user-level, beyond HWID bans). All routes
// require the admin token via adminAuthorized and are disabled when no admin
// token is configured.
package api

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/database"
)

// handleAdminListUsers handles GET /api/admin/users with search + pagination.
func (s *Server) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	accounts, err := s.db.SearchAccounts(q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	total, err := s.db.CountAccounts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	users := make([]map[string]any, 0, len(accounts))
	for _, acc := range accounts {
		mcBanned, _ := s.db.IsMinecraftBanned(acc.ID)
		approved, _ := s.db.IsAccountApproved(acc.ID)
		banned, _, _ := s.db.IsAccountBanned(acc.ID)
		muteBans, _ := s.db.ActiveBansFor("mute", strconv.FormatInt(acc.ID, 10))
		users = append(users, map[string]any{
			"id":               acc.ID,
			"username":         acc.Username,
			"created_at":       acc.CreatedAt.UTC().Format(time.RFC3339),
			"minecraft_banned": mcBanned,
			"approved":         approved,
			"banned":           banned,
			"muted":            len(muteBans) > 0,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users, "total": total})
}

// handleAdminDeleteUser handles DELETE /api/admin/users/{id}.
func (s *Server) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	acc, _ := s.db.AccountByID(id)
	target := strconv.FormatInt(id, 10)
	if acc != nil {
		target = acc.Username
	}
	if err := s.db.DeleteAccount(id); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = s.db.RecordAudit("admin", "accounts.delete", target, "fiók véglegesen törölve")
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---------------------------------------------------------------------------
// Invite allowlist management (beta registration)
// ---------------------------------------------------------------------------

func (s *Server) handleAdminListInvites(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	invites, err := s.db.ListInvites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]map[string]any, 0, len(invites))
	for _, inv := range invites {
		out = append(out, map[string]any{
			"id":         inv.ID,
			"identifier": inv.Identifier,
			"created_at": inv.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"invites": out})
}

func (s *Server) handleAdminAddInvite(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	var req struct {
		Identifier string `json:"identifier"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	inv, err := s.db.AddInvite(req.Identifier)
	if err != nil {
		if errors.Is(err, database.ErrConflict) {
			writeError(w, http.StatusConflict, "identifier already invited")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid identifier")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         inv.ID,
		"identifier": inv.Identifier,
	})
}

func (s *Server) handleAdminRemoveInvite(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid invite id")
		return
	}
	if err := s.db.RemoveInvite(id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ---------------------------------------------------------------------------
// Minecraft-level ban (distinct from system/HWID bans)
// ---------------------------------------------------------------------------

func (s *Server) handleAdminMinecraftBan(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Banned bool `json:"banned"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	if err := s.db.SetMinecraftBanned(id, req.Banned); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	status := "unbanned"
	if req.Banned {
		status = "banned"
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

// ---------------------------------------------------------------------------
// Granular role / permission management
// ---------------------------------------------------------------------------

func (s *Server) handleAdminListRoles(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	roles, err := s.db.ListRoles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]map[string]any, 0, len(roles))
	for _, rl := range roles {
		out = append(out, map[string]any{"id": rl.ID, "name": rl.Name, "description": rl.Description})
	}
	writeJSON(w, http.StatusOK, map[string]any{"roles": out})
}

func (s *Server) handleAdminListPermissions(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	perms, err := s.db.ListPermissions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]map[string]any, 0, len(perms))
	for _, p := range perms {
		out = append(out, map[string]any{"id": p.ID, "name": p.Name, "description": p.Description})
	}
	writeJSON(w, http.StatusOK, map[string]any{"permissions": out})
}

func (s *Server) handleAdminGrantRole(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Role) == "" {
		writeError(w, http.StatusBadRequest, "role required")
		return
	}
	if err := s.db.AddAdminRole(id, strings.TrimSpace(req.Role)); err != nil {
		if errors.Is(err, database.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if acc, err := s.db.AccountByID(id); err == nil {
		_ = s.db.RecordAudit("admin", "roles.assign", acc.Username, "rang: "+strings.TrimSpace(req.Role))
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "granted"})
}

func (s *Server) handleAdminRevokeRole(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	role := strings.TrimSpace(r.PathValue("role"))
	if role == "" {
		writeError(w, http.StatusBadRequest, "role required")
		return
	}
	if err := s.db.RemoveAdminRole(id, role); err != nil {
		if errors.Is(err, database.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if acc, err := s.db.AccountByID(id); err == nil {
		_ = s.db.RecordAudit("admin", "roles.revoke", acc.Username, "rang: "+role)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (s *Server) handleAdminUserRoles(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	roleRows, err := s.db.AccountRoles(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	names := make([]string, 0, len(roleRows))
	for _, rl := range roleRows {
		names = append(names, rl.Name)
	}
	writeJSON(w, http.StatusOK, map[string]any{"roles": names})
}

func (s *Server) handleAdminGrantPermission(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Permission string `json:"permission"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Permission) == "" {
		writeError(w, http.StatusBadRequest, "permission required")
		return
	}
	p, err := s.db.PermissionByName(strings.TrimSpace(req.Permission))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "permission not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	roleName := "user:" + strconv.FormatInt(id, 10)
	role, err := s.db.RoleByName(roleName)
	if err != nil {
		role, err = s.db.CreateRole(roleName, "Per-user permissions for account "+strconv.FormatInt(id, 10))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
	}
	if err := s.db.AddAccountRole(id, role.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.db.AddRolePermission(role.ID, p.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "granted"})
}

// handleAdminAccountApproval sets the approved flag for an account.
func (s *Server) handleAdminAccountApproval(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Approved bool `json:"approved"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	if err := s.db.SetAccountApproved(id, req.Approved); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "approved": req.Approved})
}

// handleAdminSkinApproval sets the approval flag for a skin file id.
func (s *Server) handleAdminSkinApproval(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid skin id")
		return
	}
	var req struct {
		Approved bool `json:"approved"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	if err := s.db.SetSkinFileApproval(id, req.Approved, "admin"); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	detail := "elutasítva"
	if req.Approved {
		detail = "jóváhagyva"
	}
	_ = s.db.RecordAudit("admin", "skins.approve", id, detail)
	writeJSON(w, http.StatusOK, map[string]any{"skin_id": id, "approved": req.Approved})
}

// handleAdminTripwires lists recent tripwire (tamper/anomaly) events.
func (s *Server) handleAdminTripwires(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	events, err := s.db.ListTripwireEvents(200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "total": len(events)})
}

// handleAdminResetPassword generates a temporary password for an account,
// stores its bcrypt hash, revokes all sessions and returns the temp password
// once (the caller must show it to the admin).
func (s *Server) handleAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	acc, err := s.db.AccountByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	// Generate a random temporary password.
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%"
	temp := make([]byte, 12)
	for i := range temp {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		temp[i] = chars[n.Int64()]
	}
	tempStr := string(temp)
	hash, err := auth.HashPassword(tempStr, s.cfg.BcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.db.UpdatePassword(id, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = s.db.RevokeAllSessions(id)
	_ = s.db.RecordAudit("admin", "accounts.password-reset", acc.Username, "ideiglenes jelszó generálva")
	writeJSON(w, http.StatusOK, map[string]any{
		"username":      acc.Username,
		"temp_password": tempStr,
	})
}

// handleAdminAudit returns the immutable audit log.
func (s *Server) handleAdminAudit(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200
	}
	entries, err := s.db.ListAudit(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"id":         e.ID,
			"actor":      e.Actor,
			"action":     e.Action,
			"target":     e.Target,
			"detail":     e.Detail,
			"created_at": e.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": out, "total": len(out)})
}

// handleAdminListPendingSkins returns skin files pending approval.
func (s *Server) handleAdminListPendingSkins(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	pending, err := s.db.ListPendingSkinFiles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]map[string]any, 0, len(pending))
	for _, p := range pending {
		out = append(out, map[string]any{
			"skin_id":     p.SkinID,
			"approved":    p.Approved,
			"reviewer":    p.Reviewer,
			"reviewed_at": p.ReviewedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"skins": out, "total": len(out)})
}
