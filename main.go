package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"reflect"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v31/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var args struct {
		Name              string `envconfig:"RUNNER_NAME"`
		Home              string `envconfig:"RUNNER_HOME" default:"/opt/runner"`
		Owner             string `envconfig:"RUNNER_OWNER"`
		Repo              string `envconfig:"RUNNER_REPO"`
		Org               string `envconfig:"RUNNER_ORG"`
		AccessToken       string `envconfig:"RUNNER_ACCESS_TOKEN"`
		AppID             int64  `envconfig:"RUNNER_APP_ID"`
		AppInstallationID int64  `envconfig:"RUNNER_APP_INSTALLATION_ID"`
		AppPrivateKey     string `envconfig:"RUNNER_APP_PRIVATE_KEY"`
		AppPrivateKeyPath string `envconfig:"RUNNER_APP_PRIVATE_KEY_PATH"`
		privateKey        []byte
	}


	err := envconfig.Process("", &args)
	if err != nil {
		log.Fatal(err)
	}
	unsetEnvironment(&args)

	var (
		c   *github.Client
		t   *github.RegistrationToken
		url string
	)

	if args.Name == "" {
		args.Name, _ = os.Hostname()
	}

	switch {
	case args.AppPrivateKey != "":
		args.privateKey = []byte(args.AppPrivateKey)
	case args.AppPrivateKeyPath != "":
		args.privateKey, err = ioutil.ReadFile(args.AppPrivateKeyPath)
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to read private key from file"))
		}
	}

	switch {
	case args.AppID != 0 && args.AppInstallationID != 0 && args.privateKey != nil:
		c, err = clientInstallation(ctx, args.AppID, args.AppInstallationID, args.privateKey)
	case args.AccessToken != "":
		c, err = clientAccessToken(ctx, args.AccessToken)
	default:
		log.Fatal("not enough data to initialize the client")
	}
	if err != nil {
		log.Fatal(errors.Wrap(err, "client initialization error"))
	}

	switch {
	case args.Owner != "" && args.Repo != "":
		t, err = tokenRepo(ctx, c, args.Owner, args.Repo)
		url = fmt.Sprintf("https://github.com/%v/%v", args.Owner, args.Repo)
	case args.Org != "" && args.Repo != "":
		t, err = tokenRepo(ctx, c, args.Org, args.Repo)
		url = fmt.Sprintf("https://github.com/%v/%v", args.Org, args.Repo)
	case args.Org != "":
		t, err = tokenOrg(ctx, c, args.Org)
		url = fmt.Sprintf("https://github.com/%v", args.Org)
	default:
		log.Fatal("not enough data to register runner")
	}
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to create registration token"))
	}

	err = execUnregister(ctx, args.Home, args.Name, url, t.GetToken())
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to unregister runner"))
	}

	err = execRegister(ctx, args.Home, args.Name, url, t.GetToken())
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to register runner"))
	}

	err = execRun(ctx, args.Home)
	if err != nil {
		log.Fatal(errors.Wrap(err, "runner failed"))
	}
}

func clientAccessToken(ctx context.Context, token string) (*github.Client, error) {
	t := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	o := oauth2.NewClient(ctx, t)
	c := github.NewClient(o)
	return c, nil
}

func clientInstallation(_ context.Context, app, installation int64, key []byte) (*github.Client, error) {
	tr, err := ghinstallation.New(http.DefaultTransport, app, installation, key)
	if err != nil {
		return nil, err
	}
	return github.NewClient(&http.Client{Transport: tr}), nil
}

func tokenRepo(ctx context.Context, c *github.Client, owner, repo string) (*github.RegistrationToken, error) {
	token, _, err := c.Actions.CreateRegistrationToken(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func tokenOrg(ctx context.Context, c *github.Client, org string) (*github.RegistrationToken, error) {
	token, _, err := c.Actions.CreateOrganizationRegistrationToken(ctx, org)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func execUnregister(ctx context.Context, home, name, url, token string) error {
	cmd := exec.CommandContext(ctx,
		path.Join(home, "config.sh"),
		[]string{
			"remove",
			"--name", name,
			"--url", url,
			"--token", token,
		}...
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func execRegister(ctx context.Context, home, name, url, token string) error {
	cmd := exec.CommandContext(ctx,
		path.Join(home, "config.sh"),
		[]string{
			"--unattended",
			"--replace",
			"--name", name,
			"--url", url,
			"--token", token,
		}...
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func execRun(ctx context.Context, home string) error {
	cmd := exec.CommandContext(ctx, path.Join(home, "run.sh"), []string{"--once",}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func unsetEnvironment(x interface{}) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		panic("should be ptr")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic("should be struct")
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, ok := f.Tag.Lookup("envconfig")
		if !ok {
			continue
		}
		_ = os.Unsetenv(name)
	}
}