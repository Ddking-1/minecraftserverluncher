package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) handleAdminKickUser(w http.ResponseWriter, r *http.Request) {
	s.handleAdminModerationAction("kick", w, r)
}

func (s *Server) handleAdminMuteUser(w http.ResponseWriter, r *http.Request) {
	s.handleAdminModerationAction("mute", w, r)
}

func (s *Server) handleAdminWarnUser(w http.ResponseWriter, r *http.Request) {
	s.handleAdminModerationAction("warn", w, r)
}

func (s *Server) handleAdminUnmuteUser(w http.ResponseWriter, r *http.Request) {
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
			_ = s.db.UnbanByTypeTarget("mute", fmt.Sprintf("%d", id))
		case "hwid":
			devs, _ := s.db.DevicesByAccountSince(id, time.Now().UTC().AddDate(0, 0, -3650))
			for _, d := range devs {
				_ = s.db.UnbanByTypeTarget("mute", d.DeviceID)
			}
		case "ip":
			ips := distinctIPsForAccount(s.db, id, 3650)
			for _, ip := range ips {
				_ = s.db.UnbanByTypeTarget("mute", ip)
			}
		}
	}
	_ = s.db.RecordAudit("admin", "players.unmute", target, "némítás feloldva")
	writeJSON(w, http.StatusOK, map[string]string{"status": "unmuted"})
}

func (s *Server) handleAdminModerationAction(actionType string, w http.ResponseWriter, r *http.Request) {
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
		Scope           string   `json:"scope"`
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
		reason = actionType + " by admin"
	}
	if len(reason) > 500 {
		writeError(w, http.StatusBadRequest, "reason too long")
		return
	}
	since := time.Now().UTC().AddDate(0, 0, -req.LookbackDays)

	counts := map[string]int{}
	affectedIPAccounts := []int64{}

	for _, t := range req.Targets {
		if t == "account" {
			aid := id
			_, _ = s.db.CreateBan(actionType, req.Scope, fmt.Sprintf("%d", id), &aid, reason, req.DurationMinutes)
			counts["account"] = 1
			if actionType == "kick" && (req.Scope == "platform" || req.Scope == "both") {
				_ = s.db.RevokeAllSessions(id)
			}
		}
	}
	for _, t := range req.Targets {
		if t == "hwid" {
			devs, _ := s.db.DevicesByAccountSince(id, since)
			n := 0
			for _, d := range devs {
				aid := id
				_, _ = s.db.CreateBan(actionType, req.Scope, d.DeviceID, &aid, reason, req.DurationMinutes)
				n++
			}
			if actionType == "kick" && (req.Scope == "platform" || req.Scope == "both") {
				_ = s.db.RevokeAllSessions(id)
			}
			counts["hwid"] = n
		}
	}
	for _, t := range req.Targets {
		if t == "ip" {
			ips := []string{}
			if strings.TrimSpace(req.IP) != "" {
				ips = append(ips, strings.TrimSpace(req.IP))
			} else {
				ips = distinctIPsForAccount(s.db, id, req.LookbackDays)
			}
			n := 0
			for _, ip := range ips {
				aid := id
				_, _ = s.db.CreateBan(actionType, req.Scope, ip, &aid, reason, req.DurationMinutes)
				n++
				accs, _ := s.db.AccountsByIP(ip, req.LookbackDays)
				affectedIPAccounts = append(affectedIPAccounts, accs...)
			}
			if actionType == "kick" && (req.Scope == "platform" || req.Scope == "both") {
				seen := map[int64]bool{}
				for _, aid := range affectedIPAccounts {
					if seen[aid] {
						continue
					}
					seen[aid] = true
					_ = s.db.RevokeAllSessions(aid)
				}
			}
			counts["ip"] = n
		}
	}

	seen := map[int64]bool{}
	var dedup []int64
	for _, a := range affectedIPAccounts {
		if !seen[a] {
			seen[a] = true
			dedup = append(dedup, a)
		}
	}

	_ = s.db.RecordAudit("admin", "players."+actionType, acc.Username, reason)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":               actionStatus(actionType),
		"account":              acc.Username,
		"counts":               counts,
		"affected_ip_accounts": dedup,
		"affected_ip_count":    len(dedup),
	})
}

func (s *Server) accountIsMuted(accountID int64) bool {
	if bans, _ := s.db.ActiveBansFor("mute", fmt.Sprintf("%d", accountID)); len(bans) > 0 {
		return true
	}
	devs, _ := s.db.DevicesByAccountSince(accountID, time.Now().UTC().AddDate(0, 0, -3650))
	for _, d := range devs {
		if bans, _ := s.db.ActiveBansFor("mute", d.DeviceID); len(bans) > 0 {
			return true
		}
	}
	for _, ip := range distinctIPsForAccount(s.db, accountID, 3650) {
		if bans, _ := s.db.ActiveBansFor("mute", ip); len(bans) > 0 {
			return true
		}
	}
	return false
}

func actionStatus(actionType string) string {
	switch actionType {
	case "mute":
		return "muted"
	default:
		return actionType + "ed"
	}
}
