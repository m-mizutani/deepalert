package deepalert

import "time"

var (
	NewReportCoordinator = newReportCoordinator
	NewReportID          = newReportID
)

func TakeReportID(x *reportCoordinator, alertID string, ts time.Time) (ReportID, error) {
	return x.takeReportID(alertID, ts)
}

func SaveAlertCache(x *reportCoordinator, reportID ReportID, alert Alert) error {
	return x.saveAlertCache(reportID, alert)
}

func FetchAlertCache(x *reportCoordinator, reportID ReportID) ([]Alert, error) {
	return x.fetchAlertCache(reportID)
}
