# Tent

Tent is a content management system that leverages a Github Repository to store its content and its history.
That allows Tent to use the [**Content as code**](https://github.com/iilab/contentascode) standard.

# Configuration

Some steps are required for using tent.

## Project

Create the repository that will store your content, by using this [page](https://github.com/new). 
Once it's created go into project **Settings** in the **Webhooks** menu.
Here you can create a new hook with the following URL 
`https://YourAppPublicDomain/api/repo/update` and default settings.

## OAuth

Tent needs an OAuth application configured in order to work. 
The application can be bound to an user (using this [form](https://github.com/settings/applications/new))
or an organization (https://github.com/organizations/YourOrg/settings/applications/new).

Once it's created **client id** and **client secret** will be available. 
You can also change several other parameters including *Authorization callback URL*:
this should be set as `https://YourAppPublicDomain/auth/callback`.

## Binary

Download the latest version of the [binary](https://github.com/securityfirst/tent/releases/latest) or build it from source.
In order to do so install [latest version](https://golang.org/dl/) of Go 
then execute `go get github.com/securityfirst/tent/tent`.
This will install it in your `$GOPATH` (that should be `~/go/bin`)

## Configuration

Now you create a configuration file named `.tent.yaml` in your `$HOME` folder using the following contents as example.

```yaml
Github:
  Handler: "awesomeorg"
  Project: "myawesomeproject"
Config:
  Id: "YOUR_CLIENT_ID"
  Secret: "YOUR_CLIENT_SECRET"
  OAuthHost: "https://YourAppPublicDomain"
  Login:
    Endpoint: "/auth/login"
  Logout:
    Endpoint: "/auth/logout"
  Callback:
    Endpoint: "/auth/callback"
  RandomString: "whatever"
Port: 80
Transifex:
  Project: "project-name"
  Username: "user"
  Password: "password"
```

# Run

Once everything is ready you can start the app using `tent.exe run`.

# Sample Repo

This a sample repo used by tent: https://github.com/securityfirst/tent-content
