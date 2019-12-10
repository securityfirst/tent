# Tent

Tent is a content management system that leverages a Github Repository to store its content and its history.
That allows Tent to use the [**Content as code**](https://github.com/iilab/contentascode) standard.

# Configuration

Some steps are required for using tent.

# Server
Tent is best run on the following infrastructure:

-Ubuntu Server 16.04.3 LTS

-Windows 2012

We recommend it running with a minimum of 1GB RAM, 1 GB of storage.

# Services
Location, local laws and choice of service provider is important when considering the security of any deployment. 

For NGOs, we particularly recommend the use https://eclips.is - run by a privacy friendly provider who provide free server space in multiple jurisdictions.

# Security
We highly recommend hardening of any server that Tent is running on.

For example for Linux we recommend using the following guides:

[Ubuntu Server Hardening Guide: Quick and Secure]
(https://linux-audit.com/ubuntu-server-hardening-guide-quick-and-secure/)

[Best practices for hardening a new server]
(https://www.digitalocean.com/community/questions/best-practices-for-hardening-new-sever-in-2017)

For Windows 2012 we recommend using the following guides:

[Windows Sever 2012 Hardening Checklist]
(https://wikis.utexas.edu/display/ISO/Windows+Server+2012+R2+Hardening+Checklist)

[Server Hardening: Windows Server 2012]
(https://technet.microsoft.com/en-us/security/jj720323.aspx)

We also recommend the usage of server auditing measurements such as those provided by the [Center for Internet Security](Center for Internet Security) and [SCAP](https://scap.nist.gov/). 

Auditing tools we recommend include CIS-CAT, [Open Scap](https://www.open-scap.org) and [Lynis](https://cisofy.com/lynis/). For non-profits we also want to highlight the availability of a number of discount or free tools that can be used to help ensure the security of a Tent deployment:

[Tenable Nessus Non-Profit Donation](https://www.tenable.com/about-tenable/tenable-in-the-community/tenable-charitable-organization-subscription-program)

[Crowdstrike Non-Profit Donation]
(https://www.crowdstrike.org)

For enhanced security of the Tent server, we recommend the conducting of a penetration test according to the [SAFETAG](https://safetag.org/) methodology - designed specifically for civil society groups at risk.



## Project

Create the repository that will store your content, by using this [page](https://github.com/new). 
Once it's created go into project **Settings** in the **Webhooks** menu.
Here you can create a new hook with the following URL 
`https://YourAppPublicDomain/api/repo/update` and default settings.

## OAuth

Tent needs an OAuth application configured in order to work. 
The application can be bound to an user (using this [form](https://github.com/settings/applications/new))
or an organization (`https://github.com/organizations/YourOrg/settings/applications/new`).

Once it's created **client id** and **client secret** will be available. 
You can also change several other parameters including *Authorization callback URL*:
this should be set as `https://YourAppPublicDomain/auth/callback`.
You will need these for the configuration file.

## Binary

Download the latest version of the [binary](https://github.com/securityfirst/tent/releases/latest) or build it from source.
In order to do so install [latest version](https://golang.org/dl/) of Go 
then execute `go get github.com/securityfirst/tent/tent`.
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
  Port: 80                                  # Port used by the App
Transifex:
  Project: "project-name"
  Username: "user"
  Password: "password"
  Language: en
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

This a the repo used by tent in the [Umbrella App](https://play.google.com/store/apps/details?id=org.secfirst.umbrella): https://github.com/securityfirst/tent-content
This is another sample repo, with sample content, that you can use to check how things works: https://github.com/securityfirst/tent-example

# New! Branches

You can work on a specific branch of your content project, this is very usefull for testing purpose (i.e. a big update on content).

# Troubleshot

When you execute `tent run` you can check if the app does the checkout correctly:

```
Checkout with 5ecb0438dc9aec32e4f5d4572584cd9c50b53055 refs/remotes/origin/master
```

If some files is not formatted correctly and raises an error you will see something like:

```
Parsing failed: [contents]filepath/file.md - Invalid content
```

# Localisation

You can upload the app content to transifex for translation with the command `tent transifex upload`

After translations are done you can use `tent transifex download` to import them.
