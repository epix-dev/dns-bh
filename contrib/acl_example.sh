#!/bin/sh

vtysh -c 'show ip bgp community xxx:yyy' \
    | awk '/^\*.?.?[0-9]/{ gsub("\*.?.?", "", $0); gsub("\.0$", ".0/24", $1); print $1 }' \
    | uniq \
    | sort \
    | /opt/dns-bh/bin/acl

if [ -s /opt/dns-bh/etc/acl.txt ]; then
    if ! cmp --quiet /opt/dns-bh/etc/acl.txt /etc/powerdns/pdns_acl.txt; then
        cp /opt/dns-bh/etc/acl.txt /etc/powerdns/pdns_acl.txt
        rec_control > /dev/null reload-acls
    fi
fi
