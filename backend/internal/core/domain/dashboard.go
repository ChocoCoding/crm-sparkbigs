package domain

// DashboardStats agrega las métricas globales del CRM para un usuario.
type DashboardStats struct {
	// KPIs numéricos
	TotalCompanies      int64   `json:"total_companies"`
	ActiveCompanies     int64   `json:"active_companies"`
	TotalContacts       int64   `json:"total_contacts"`
	ActiveSubscriptions int64   `json:"active_subscriptions"`
	MRR                 float64 `json:"mrr"`  // suma suscripciones activas mensuales
	ARR                 float64 `json:"arr"`  // suma suscripciones activas anuales
	MeetingsThisMonth   int64   `json:"meetings_this_month"`

	// Listas resumidas
	UpcomingMeetings []Meeting      `json:"upcoming_meetings"`
	ExpiringSoon     []Subscription `json:"expiring_soon"`
	RecentCompanies  []Company      `json:"recent_companies"`
}
