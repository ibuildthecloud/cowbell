package server

import "html/template"

const (
	info = `
Available web hooks receivers

Scale up

curl -X POST http://server/{token}/scale/{stack}/{service}/up

Scale down

curl -X POST http://server/{token}/scale/{stack}/{service}/down

Scale to X

curl -X POST http://server/{token}/scale/{stack}/{service}/1

Upgrade

curl -X POST http://server/{token}/upgrade/{stack}?dockerComposeUrl=http://...&rancherComposeUrl=http://...
curl -X POST http://server/{token}/upgrade/{stack}/{service}?dockerComposeUrl=http://...&rancherComposeUrl=http://...

Redeploy

curl -X POST http://server/{token}/redeploy/{stack}
curl -X POST http://server/{token}/redeploy/{stack}/{service}

List jobs

curl http://server/{token}/jobs
`
	list = `
<html>
	<table>
			<tr>
				<td>
					ID
				</td>
				<td>
					State
				</td>
				<td>
					Output Link
				</td>
				<td>
					Created
				</td>
			</tr>
		{{range .jobs}}
			<tr>
				<td>
					{{.ID}}
				</td>
				<td>
					{{.State}}
				</td>
				<td>
					<a href="jobs/{{.ID}}/output">Output</a>
				</td>
				<td>
					{{.Created}}
				</td>
			</tr>
		{{end}}
	</table>
</html>
`
)

var (
	infoTemplate = template.Must(template.New("info").Parse(info))
	listTemplate = template.Must(template.New("list").Parse(list))
)
