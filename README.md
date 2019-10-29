# ScanBanServer
The server which handles all IP reports and provides a list with all evil IP addresses

# Dynamic IP
If your server has a dynamic IP address, you can create a dyn.ip file and let a cronjob update this file with your external ip. This prevents the server reporting itself after ip change. If the file is avaliable and contains a valid IP address, it will be used instead of the external ip.
