# cloud-repl

A simulated terminal to provide tours of Rackspace Cloud to developers.

# Setup

### 1. Required installations

Install Docker, git and vim

### 2. Specify API creds

Copy the `env.list.dist` specified in this repo, and fill in your own credentials.

### 3. Install and run Go API

```bash
docker pull jamiehannaford/provisioner
docker run -d -p 8080:8080 --name provisioner --env-file ./env.list jamiehannaford/provisioner:latest
```

### 4. Install and run HTML frontend

```bash
docker pull jamiehannaford/frontend
docker run -d -p 80:80 --name frontend jamiehannaford/frontend:latest
```
