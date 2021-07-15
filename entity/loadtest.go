package entity

type LoadTestMessage struct {
	Type string `json:"type"`
}

type LoadTestConfig struct {
	Domain    string  `json:"domain"`
	Initial   int     `json:"initial"`
	Increment int     `json:"increment"`
	Payload   Payload `json:"payload"`
}

type LoadTestResult struct {
	Message           string  `json:"message"`
	SuccessCount      int     `json:"success_count"`
	FailureCount      int     `json:"failure_count"`
	TotalRequestCount int     `json:"total_request_count"`
	Rps               int     `json:"rps"`
	Elapsed           float64 `json:"elapsed"`
	FastestTime       float64 `json:"fastest_time"`
	SlowestTime       float64 `json:"slowest_time"`
}

type Request struct {
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
	Body     interface{}       `json:"body"`
}

type Payload struct {
	Override struct {
		Headers map[string]string `json:"headers"`
	} `json:"override"`
	Requests []Request `json:"requests"`
}
