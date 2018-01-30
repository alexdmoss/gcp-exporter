package commands

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/client"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors"
)

var (
	getTokenOutputTpl = `
      Token: {{.AccessToken}}
       Type: {{.TokenType}}
 Expires at: {{.Expiry}}

 ----
 example usage: curl -s -X GET \
                     -H "Authorization: {{.TokenType}} {{.AccessToken}}" \
                     https://www.googleapis.com/compute/v1/projects/[PROJECT]/zones/[ZONE]/instances
`
)

type GetTokenCommand struct {
	ServiceAccountFile string `long:"service-account-file" env:"GCP_EXPORTER_SERVICE_ACCOUNT_FILE" description:"Path to GCP Service Account JSON file"`
}

func (gtc *GetTokenCommand) Execute(*cli.Context) {
	ts := client.NewGCPServiceAccountTokenSource(gtc.ServiceAccountFile)

	token, err := ts.Token()
	if err != nil {
		logrus.WithError(err).Fatalln("error while requesting new token")
	}

	tpl, err := template.New("output").Parse(getTokenOutputTpl)
	if err != nil {
		logrus.WithError(err).Fatalln("error while parsing template")
	}

	buff := bytes.NewBufferString("")
	tpl.Execute(buff, token)

	fmt.Println(buff.String())
}

func NewGetTokenCommand() cli.Command {
	cmd := &GetTokenCommand{
		ServiceAccountFile: collectors.DefaultServiceAccountFile,
	}

	return PrepareCommand("get-token", "Request the oAuth2 Token from GCP", cmd)
}
