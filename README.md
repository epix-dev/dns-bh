# DNS-BH

Proposal of simple DNS blackhole/sinkhole solution based on custom software written in Go and PowerDNS Recursor + LUA plugin.

The solution is intended for interaction with the Polish system: http://hazard.mf.gov.pl and Malware Domains Blocklist (http://www.malwaredomains.com).

# INSTALL

Example install procedure based on Ubuntu 16.04 LTS.

Download and unpack tarball from: https://github.com/epix-dev/dns-bh/releases/latest

## Master controller host

This host provides database, PULL/PUSH daemon for http://hazard.mf.gov.pl and domains fetcher for http://www.malwaredomains.com.

Auth by Client Certificate and SSL traffic handle are provided by NGINX which is reverse proxy for PULL/PUSH daemon.

Install required packages:

    apt-get install nginx-light postgresql

Configure PostgreSQL, create dns-bh_production database owned by dns-bh user and load schema from build/contrib/schema.sql.

Check certificate fingerprint, for example:

    openssl x509 -in pl.gov.mf.hazard.crt -fingerprint -noout \
        | cut -d= -f2 \
        | tr -d : \
        | tr '[:upper:]' '[:lower:]'


Configure required HTTP proxy for handling SSL traffic:

    cp build/contrib/nginx/dns-bh /etc/nginx/sites-available
    ln -s /etc/nginx/sites-available/dns-bh /etc/nginx/sites-enabled/

Edit required options in /etc/nginx/sites-available/dns-bh and finally:

    systemctl restart nginx.service

### Install dns-bh software

Install dns-bh files:

    cp -r build/dns-bh_master /opt/dns-bh
    chown -R nobody:nogroup /opt/dns-bh

Install systemd services:

    cp build/contrib/systemd/dns-bh_hazard.* /etc/systemd/system
    cp build/contrib/systemd/dns-bh_malware.* /etc/systemd/system

    systemctl enable dns-bh_hazard.service
    systemctl start dns-bh_hazard.service

    systemctl enable dns-bh_malware.timer
    systemctl start dns-bh_malware.timer

Example crontab entry which force PULL request every day at midnight

    0   0 * * *     root    killall -HUP hazard

Configure database connection and SMTP settings in /opt/dns-bh/config.yml

## Resolver nodes

Node host is DNS resolver based on PowerDNS Recursor, LUA plugin and simple domains exporter.

Repeat this steps on all nodes.

### Install dns-bh software

Install dns-bh files:

    cp -r build/dns-bh_node /opt/dns-bh
    chown -R nobody:nogroup /opt/dns-bh

Install systemd services:

    cp build/contrib/systemd/dns-bh_export-file.* /etc/systemd/system

    systemctl enable dns-bh_export-file.timer
    systemctl start dns-bh_export-file.timer

Create the file '/etc/apt/sources.list.d/pdns.list' with this content:

    deb [arch=amd64] http://repo.powerdns.com/ubuntu xenial-auth-40 main

And this content to '/etc/apt/preferences.d/pdns':

    Package: pdns-*
    Pin: origin repo.powerdns.com
    Pin-Priority: 600

Execute the following commands:

    curl https://repo.powerdns.com/FD380FBB-pub.asc | apt-key add -
    apt-get update
    apt-get install pdns-recursor

For details see: https://repo.powerdns.com/

Copy LUA script to powerdns config directory:

    cp build/contrib/powerdns/* /etc/powerdns

Configure pdns-recursor, enable and customize LUA script.

Example crontab entry which check for dns-bh.reload file and reloading lua script without restart pdns

    */5 * * * *     root    test -f /etc/powerdns/dns-bh.reload && (rec_control reload-lua-script; rm /etc/powerdns/dns-bh.reload)

Configure database connection /opt/dns-bh/config.yml, SMTP settings is not used.

# LICENSE

My sources are licensed under MIT license, other sources might have it's own licenses.

# RESOURCES

- https://doc.powerdns.com/md/recursor/scripting/#helpful-functions
- https://github.com/PowerDNS/pdns/blob/master/pdns/powerdns-example-script.lua
- http://www.finanse.mf.gov.pl/inne-podatki/podatek-od-gier-gry-hazardowe/komunikaty/-/asset_publisher/d3oA/content/konsultacje-specyfikacji-technicznej-interfejsu-oraz-specyfikacji-wejscia-wyjscia-dla-rejestru-domen-sluzacych-do-oferowania-gier-hazardowych-niezgodnie-z-ustawa
