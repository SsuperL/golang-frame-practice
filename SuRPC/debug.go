package surpc

import (
	"fmt"
	"html/template"
	"net/http"
)

var debugText = `<html>
<body>
<title>SuRPC Services</title>
{{range .}}
<hr>
Service {{.Name}}
<hr>
	<table>
	<th align=center>Method</th><th align=center>Calls</th>
	{{range $name, $mtype := .Method}}
		<tr>
		<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
		<td align=center>{{$mtype.NumCalls}}</td>
		</tr>
	{{end}}
	</table>
{{end}}
</body>
</html>`

var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugService struct {
	Name   string
	Method map[string]*methodType
}

type debugHTTP struct {
	*Server
}

func (server debugHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var services []debugService
	server.serviceMap.Range(func(namei, svci interface{}) bool {
		svc := svci.(*service)
		services = append(services, debugService{
			Name:   namei.(string),
			Method: svc.method,
		})
		return true
	})
	err := debug.Execute(w, services)
	if err != nil {
		fmt.Fprintln(w, "rpc: error excuting template: ", err.Error())
	}
}
