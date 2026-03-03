package data

type Appointment struct {
	ID    int    `json:"id"`
	BusinessID int    `json:"business_id"`
	BusinessName string `json:"business_name,omitempty"`
	ServiceID int    `json:"service_id"`
	ServiceName string `json:"service_name,omitempty"`
	BusinessStaffID int    `json:"business_staff_id"`
	BusinessStaffName string `json:"business_staff_name,omitempty"`
	CustomerID int    `json:"customer_id"`
	CustomerName string `json:"customer_name,omitempty"`
	Date  string `json:"date"`
}