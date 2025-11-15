I am deploying this website to my personal linode, and taking some notes on what I did.

I am using the domain name potbot.online and the linode at IP address 192.46.218.92 (same linode that hosts TSLC and my personal website).

# Steps I took

## DNS

- Purchased domain name `potbot.online` for 1 dollar for 1 year (that's a good deal!)
- Updated namecheap to have a redirect record "www.potbot.online" -> "https://potbot.online"
- Updated namecheap to have an A record that matches A record format of eric289.com, but with automatic TTL.

## Repo setup

Cloned the repo into directory `~/potbot`

ran `npm install` and `npm run build` in the front end folder (I already have npm installed)

installed go version 1.25 by running 
```bash
wget https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.4.linux-amd64.tar.gz
```

added the following info in the `.env` file in the backend directory

```text
DATABASE_USER=xxxxxx
DATABASE_PASSWORD=xxxxxx
POTBOT_HASH_KEY=xxxxxxxxxxxxxxxx
POTBOT_BLOCK_KEY=xxxxxxxxxxxxxxxx
PORT=8180
```

## Firewall
**Temporary** changes were made to the firewall for testing

- ran `sudo ufw status` to see rules.
- ran `sudo ufw allow 8180/tcp`.
	- This was a temporary thing for testing, since later apache will accept incoming connections on typical HTTPS port (443)

ran `go run .` in the backend directory just to test. Note that the backend will serve the frontend at the root, and serve the api at `/api`.

At this point, I went to http://192.46.218.92:8180/api/ping in my browser, which returned "pong". I also see the potbot login screen at http://192.46.218.92:8180. Yay!


## Mysql

On my local laptop, generated a sql dump file, then copied it to the server (I already had mysql installed on the server). I did an ugly hack when generating the sql dump (for compatibility to older mysql versions).

Locally, ran:
```
mysqldump -u root -p potbot | sed 's/utf8mb4_0900_ai_ci/utf8mb4_unicode_ci/g' > dump.sql
scp dump.sql eric:~ 
trash dump.sql
```
On the remote, ran:
```
mysql -u ericwu -p -e "CREATE DATABASE potbot;"
mysql -u ericwu -p potbot < ~/dump.sql
```

Verify that when going to http://192.46.218.92:8180 on the browser, I can now login with my credentials!

## SSL cert and  Apache web server

copied the following contents into `/etc/apache2/sites-available/potbot.online.conf`

```
<VirtualHost *:80>
	ServerName potbot.online

	ServerAdmin eric.dianhao.wu@gmail.com

	Redirect permanent / https://potbot.online

	ErrorLog ${APACHE_LOG_DIR}/error.log
	CustomLog ${APACHE_LOG_DIR}/access.log combined

	<Location />
		ProxyPass http://127.0.0.1:8180/
		ProxyPassReverse http://127.0.0.1:8180/
	</Location>
</VirtualHost>
```

ran
```
sudo a2ensite potbot.online
sudo systemctl reload apache2
sudo certbot --apache 
```

and picked the option for "potbot.online".
Got a success message.

Now, after running `go run .`, I can navigate to `potbot.online` in my browser and see the site, yay!

ran `sudo ufw delete allow 8180/tcp` to get rid of that temporary allow firewall rule, not necessary anymore since I wont be going to `http://192.46.218.92:8180` in my browser anymore.

## Creating a service with systemctl

My service for the TSLC backend is at path `/etc/systemd/system/tswiftrs.service`

copied this file to `potbot.service` and changed stuff.

Ran `go build` in the backend directory to create an executable.

ran
```
sudo systemctl enable potbot
sudo systemctl start potbot
sudo systemctl status potbot
```

## Bash script for updating the site easily

I also created a `run_build.sh` in the root of this repo, for easily doing the following steps:

- build the client (npm)
- build the server (go)
- restart the systemctl service



