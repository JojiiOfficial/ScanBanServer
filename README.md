### !This is currently in development!
<br>

# ScanBanServer
This is a REST-API server for the [tripwire](https://github.com/JojiiOfficial/Tripwire-reporter) system. If you own multiple server on the internet which get scanned all over the time, this is the perfect ip-banning-network for you. It is created to log portscans and "report" them to your REST server(this repository on one of your server). If you want to block those scanner you can automatically generate IP-blocklists based on a filter to protect your machines against internet scanner like [shodan.io](https://shodan.io)<br>

# What!? I need an example!
Lets explain this repo with a short example:<br>
Lets assume you have 5 server but you don't want known-portscanner to scan all your ports and services and publish the data on their websites, you can install [tripwire](https://github.com/JojiiOfficial/Tripwire) and [tripwire reporter](https://github.com/JojiiOfficial/Tripwire-reporter) on all of your server and configure them to report all IPs which tried to connect to port x (with tripwire you can also use ports which are already bound to a diffirent service). Then you can configue that tripwire fetches all IPs from your ScanBanServer which were for example reported at least two times. So if machine 1 and 2 were scanned by ip x.x.x.x then machine 3,4 and 5 will block the ip x.x.x.x (Of course you can create a filter matching your imagination)<br>

# Install
- Install go1.13
- Setup a mysql server (you can also use an existing mysql server)
- Create a new database 
- Import the database.sql from this repo
- Build the server binary wih `go get` and `go build -o scanban`
- Run `./scanban install`
- Fill the config.json with the database credentials (you can leave "keyfile" and "certfile" empty if you don't need TLS)
  - If you want to fetch some third party data you can set the "ipdataAPIkey" with an API key from [ipdata.co](https://ipdata.co/)
- Run `./scanban install` again to automatically cerate a systemd service
  - Alternatively you can use `./scanban run` to run the server without a service
- Done!

## Dynamic IP
If your server has a dynamic IP address, you can create a dyn.ip file and let a cronjob update this file with your external ip. This prevents the server reporting itself after ip change. If the file is avaliable and contains a valid IP address, it will be used instead of the external ip.
