package pagereader

import (
	"fmt"
	"time"
)

type Notify struct {
	FunctionName string
	MarkName     string
	StartingTime time.Time
	EndTime      time.Time
	Logs         []string
	Error        error
}

func NewNotify(functionName, markName string) *Notify {
	return &Notify{
		FunctionName: functionName,
		MarkName:     markName,
		StartingTime: time.Now(),
		Logs:         make([]string, 0),
	}
}

func (n *Notify) AddLog(msg string) *Notify {
	n.Logs = append(n.Logs, fmt.Sprintf("%s > %s", time.Now().Format("2006-01-02 15:04:05"), msg))
	return n
}

func (n *Notify) AddLogf(format string, v ...interface{}) *Notify {
	n.Logs = append(n.Logs, fmt.Sprintf(time.Now().Format("2006-01-02 15:04:05")+" > "+format, v...))
	return n
}

func (n Notify) String() string {
	format := `
Function Name: %s
    Mark Name: %s
Starting Time: %s
     End Time: %s
      Seconds: %.2f seconds
      Success: %v`
	if n.EndTime.IsZero() {
		n.EndTime = time.Now()
	}
	seconds := n.EndTime.Sub(n.StartingTime).Seconds()
	values := []interface{}{n.FunctionName, n.MarkName, n.StartingTime, n.EndTime, seconds, n.Error == nil}
	if n.Error != nil {
		format += `
       Error: %s`
		values = append(values, n.Error)
	}
	if len(n.Logs) > 0 {
		format += `
         Logs:
%s`
		logs := ""
		for _, log := range n.Logs {
			logs += "               " + log + "\n"
		}
		values = append(values, logs)
	}
	return fmt.Sprintf(format, values...)
}
