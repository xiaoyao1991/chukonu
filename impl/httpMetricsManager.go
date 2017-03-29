package impl

import "github.com/xiaoyao1991/chukonu/core"

type HttpMetricsManager struct {
}

func (m *HttpMetricsManager) RecordRequest(request core.ChukonuRequest) {
}
func (m *HttpMetricsManager) RecordResponse(response core.ChukonuResponse) {
}
func (m *HttpMetricsManager) RecordError(err error) {
}
