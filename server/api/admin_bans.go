package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"minecraftserverluncher/server/database"
)

func (s *Server) handleAdminBanUser(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		DurationMinutes int      `json:"duration_minutes"`
		Reason          string   `json:"reason"`
		Scope           string   `json:"scope"` // mc, platform, both
		Targets         []string `json:"targets"`
		LookbackDays    int      `json:"lookback_days"`
		IP              string   `json:"ip"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Scope == "" {
		req.Scope = "both"
	}
	if req.Scope != "mc" && req.Scope != "platform" && req.Scope != "both" {
		writeError(w, http.StatusBadRequest, "invalid scope")
		return
	}
	if len(req.Targets) == 0 {
		req.Targets = []string{"account"}
	}
	for _, t := range req.Targets {
		if t != "account" && t != "hwid" && t != "ip" {
			writeError(w, http.StatusBadRequest, "invalid target")
			return
		}
	}
	if req.DurationMinutes < 0 {
		writeError(w, http.StatusBadRequest, "invalid duration")
		return
	}
	if req.LookbackDays <= 0 {
		req.LookbackDays = 30
	}
	if req.LookbackDays > 365 {
		writeError(w, http.StatusBadRequest, "invalid lookback_days")
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "banned by admin"
	}
	if len(reason) > 500 {
		writeError(w, http.StatusBadRequest, "reason too long")
		return
	}
	since := time.Now().UTC().AddDate(0, 0, -req.LookbackDays)

	counts := map[string]int{}
	affectedIPAccounts := []int64{}

	// account ban
	for _, t := range req.Targets {
		if t == "account" {
			aid := id
			_, _ = s.db.CreateBan("account", req.Scope, fmt.Sprintf("%d", id), &aid, reason, req.DurationMinutes)
			counts["account"] = 1
			if req.Scope == "mc" || req.Scope == "both" {
				_ = s.db.SetMinecraftBanned(id, true)
			}
			if req.Scope == "platform" || req.Scope == "both" {
				_ = s.db.RevokeAllSessions(id)
			}
		}
	}
	// hwid ban: all devices seen in lookback
	for _, t := range req.Targets {
		if t == "hwid" {
			devs, _ := s.db.DevicesByAccountSince(id, since)
			n := 0
			for _, d := range devs {
				aid := id
				_, _ = s.db.CreateBan("hwid", req.Scope, d.DeviceID, &aid, reason, req.DurationMinutes)
				if req.Scope == "platform" || req.Scope == "both" {
					_ = s.db.BanHWID(d.DeviceID)
				}
				n++
			}
			if req.Scope == "platform" || req.Scope == "both" {
				_ = s.db.RevokeAllSessions(id)
			}
			counts["hwid"] = n
		}
	}
	// ip ban
	for _, t := range req.Targets {
		if t == "ip" {
			ips := []string{}
			if strings.TrimSpace(req.IP) != "" {
				ips = append(ips, strings.TrimSpace(req.IP))
			} else {
				// all ips this account used in lookback
				ips = distinctIPsForAccount(s.db, id, req.LookbackDays)
			}
			n := 0
			for _, ip := range ips {
				aid := id
				_, _ = s.db.CreateBan("ip", req.Scope, ip, &aid, reason, req.DurationMinutes)
				n++
				// affected accounts
				accs, _ := s.db.AccountsByIP(ip, req.LookbackDays)
				affectedIPAccounts = append(affectedIPAccounts, accs...)
			}
			if req.Scope == "platform" || req.Scope == "both" {
				for _, aid := range affectedIPAccounts {
					_ = s.db.RevokeAllSessions(aid)
				}
			}
			counts["ip"] = n
		}
	}

	// dedup affected accounts
	seen := map[int64]bool{}
	var dedup []int64
	for _, a := range affectedIPAccounts {
		if !seen[a] {
			seen[a] = true
			dedup = append(dedup, a)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":               "banned",
		"account":              acc.Username,
		"counts":               counts,
		"affected_ip_accounts": dedup,
		"affected_ip_count":    len(dedup),
	})
}

func distinctIPsForAccount(db *database.DB, accountID int64, days int) []string {
	since := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339Nano)
	rows, err := db.SQL().Query(
		`SELECT DISTINCT ip FROM account_login_ips WHERE account_id = ? AND seen_at >= ?`,
		accountID, since,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err == nil {
			out = append(out, ip)
		}
	}
	return out
}

func (s *Server) handleAdminUnbanUser(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Targets []string `json:"targets"`
	}
	_ = decodeBody(w, r, &req)
	if len(req.Targets) == 0 {
		req.Targets = []string{"account", "hwid", "ip"}
	}
	for _, t := range req.Targets {
		switch t {
		case "account":
			_ = s.db.UnbanAccount(id)
			_ = s.db.SetMinecraftBanned(id, false)
		case "hwid":
			devs, _ := s.db.DevicesByAccountSince(id, time.Now().UTC().AddDate(0, 0, -3650))
			for _, d := range devs {
				_ = s.db.UnbanByTypeTarget("hwid", d.DeviceID)
			}
		case "ip":
			ips := distinctIPsForAccount(s.db, id, 3650)
			for _, ip := range ips {
				_ = s.db.UnbanByTypeTarget("ip", ip)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unbanned"})
}

func (s *Server) handleAdminUserDetails(w http.ResponseWriter, r *http.Request) {
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
	devs, _ := s.db.DevicesByAccountSince(id, time.Now().UTC().AddDate(0, 0, -30))
	devOut := make([]map[string]any, 0, len(devs))
	for _, d := range devs {
		devOut = append(devOut, map[string]any{
			"id":        d.ID,
			"device_id": d.DeviceID,
			"name":      d.DeviceName,
			"last_seen": d.LastSeen.UTC().Format(time.RFC3339),
		})
	}
	ips := distinctIPsForAccount(s.db, id, 30)
	ipOut := make([]map[string]any, 0, len(ips))
	for _, ip := range ips {
		n, _ := s.db.CountAccountsByIP(ip, 30)
		ipOut = append(ipOut, map[string]any{"ip": ip, "accounts_last_30d": n})
	}
	bans, _ := s.db.ListBans(500)
	userBans := []map[string]any{}
	for _, b := range bans {
		if b.AccountID != nil && *b.AccountID == id {
			userBans = append(userBans, map[string]any{
				"id": b.ID, "type": b.BanType, "scope": b.Scope, "target": b.Target,
				"reason": b.Reason, "expires_at": b.ExpiresAt,
			})
		}
	}
	mcBanned, _ := s.db.IsMinecraftBanned(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":               acc.ID,
		"username":         acc.Username,
		"created_at":       acc.CreatedAt.UTC().Format(time.RFC3339),
		"minecraft_banned": mcBanned,
		"devices_last_30d": devOut,
		"ips_last_30d":     ipOut,
		"bans":             userBans,
	})
}

func (s *Server) handleAdminIPAccounts(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	if ip == "" {
		writeError(w, http.StatusBadRequest, "ip required")
		return
	}
	ids, _ := s.db.AccountsByIP(ip, 30)
	out := make([]map[string]any, 0, len(ids))
	for _, aid := range ids {
		acc, err := s.db.AccountByID(aid)
		if err == nil {
			out = append(out, map[string]any{"id": acc.ID, "username": acc.Username})
		}
	}
	n, _ := s.db.CountAccountsByIP(ip, 30)
	writeJSON(w, http.StatusOK, map[string]any{"ip": ip, "accounts": out, "count": n})
}

func (s *Server) handleAdminListBans(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200
	}
	bans, err := s.db.ListBans(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]map[string]any, 0, len(bans))
	for _, b := range bans {
		item := map[string]any{
			"id":               b.ID,
			"type":             b.BanType,
			"scope":            b.Scope,
			"target":           b.Target,
			"reason":           b.Reason,
			"duration_minutes": b.DurationMinutes,
			"expires_at":       b.ExpiresAt,
			"created_at":       b.CreatedAt.UTC().Format(time.RFC3339),
			"active":           b.IsActive(),
		}
		if b.AccountID != nil {
			item["account_id"] = *b.AccountID
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"bans": out, "total": len(out)})
}

func (s *Server) handleAdminDeleteBan(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid ban id")
		return
	}
	// Find the ban to remove related legacy entries where applicable.
	bans, _ := s.db.ListBans(10000)
	for _, b := range bans {
		if b.ID == id {
			if b.BanType == "hwid" {
				_ = s.db.UnbanHWID(b.Target)
			}
			break
		}
	}
	if _, err := s.db.SQL().Exec(`DELETE FROM bans WHERE id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unbanned"})
}

// accountHasMCBan reports whether an account is blocked from Minecraft by an
// account, HWID or IP ban with mc/both scope.
func (s *Server) accountHasMCBan(accountID int64) bool {
	if banned, _, scope := s.db.IsAccountBanned(accountID); banned && (scope == "mc" || scope == "platform" || scope == "both") {
		return true
	}
	devs, _ := s.db.DevicesByAccountSince(accountID, time.Now().UTC().AddDate(0, 0, -3650))
	for _, d := range devs {
		if banned, _, scope := s.db.IsHWIDBannedNew(d.DeviceID); banned && (scope == "mc" || scope == "platform" || scope == "both") {
			return true
		}
		if legacy, _ := s.db.IsHWIDBanned(d.DeviceID); legacy {
			return true
		}
	}
	for _, ip := range distinctIPsForAccount(s.db, accountID, 3650) {
		if banned, _, scope := s.db.IsIPBanned(ip); banned && (scope == "mc" || scope == "platform" || scope == "both") {
			return true
		}
	}
	return false
}
