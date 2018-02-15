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
You will need these for the configuration file.

## Binary

Download the latest version of the [binary](https://github.com/securityfirst/tent/releases/latest) or build it from source.
In order to do so install [latest version](https://golang.org/dl/) of Go 
then execute `go get gopkg.in/securityfirst/tent.v2/tent`.
This will install it in your `$GOPATH` (that should be `~/go/bin`)

## Configuration

Now you create a configuration file named `.tent.yaml` in your `$HOME` folder using the following contents as example.

```yaml
Github:
  Handler: "awesomeorg"                     # Github user of the project
  Project: "myawesomeproject"               # Project name
  Branch: "master"                          # Project branch (default is master)
Config:
  Id: "YOUR_CLIENT_ID"                      # replace with your client_id
  Secret: "YOUR_CLIENT_SECRET"              # replace with your secret
  OAuthHost: "https://YourAppPublicDomain"  # replace with the endpoint you specified
  Login:
    Endpoint: "/auth/login"
  Logout:
    Endpoint: "/auth/logout"
  Callback:
    Endpoint: "/auth/callback"
  State: "whatever"
Server:  
  Port: 80
Transifex:
  Project: "project-name"
  Username: "user"
  Password: "password"
```

# Run

Once everything is ready you can start the app using `tent.exe run`. You can also specify the `--config file` if you want to use a specific configuration.  

# Repo structure

The repo have the following structure

## Assets

`assets` is a folder in the project root, containing binary files (ie *pictures*) for your project. 
They are available at the `/api/repo/asset/id` endpoint.

## Forms

`forms_xx` (ie *forms_en*) is a folder containing forms file. Supports localisation (you can put a translated version in *forms_es* for instance) and has the following structure:

```md
[Name]: # (Title of the form)

[Type]: # (screen)
[Name]: # (Name of the first screen)

[Type]: # (text_area)
[Name]: # (input_name)
[Label]: # (Label of the input)

[Type]: # (screen)
[Name]: # (Another scree)

[Type]: # (multiple_choice)
[Name]: # (another-input)
[Label]: # (Label for the input)
[Options]: # (Option 1;Option 2;Option 3)
```

## Content

`content_xx` is the main content directory and it support localisation, same as forms. 
It contains 3 level of hierachy (Category, Subcategory, Difficulty) and inside a Checklist and a series of Items

```
- contents_en
  - category_id
    - .metadata.md # Category descriptor
    - subcategory_id
      - .metadata.md # Subcategory descriptor
      - difficulty_id
        - .checks.md    # Checklist
        - .metadata.md  # Difficulty descriptor
        - item_1.md     # Item
```

# Sample Repo

This a the repo used by tent in the [Umbrella App](https://play.google.com/store/apps/details?id=org.secfirst.umbrella): https://gopkg.in/securityfirst/tent.v2-content
This is another sample repo, with sample content, that you can use to check how things works: https://github.com/securityfirst/tent-example

# New! Branches

You can work on a specific branch of your content project, this is very usefull for testing purpose (i.e. a big update on content).
