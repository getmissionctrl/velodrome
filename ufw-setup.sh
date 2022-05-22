#!/bin/sh

cfufw_deleted=0
cfufw_created=0
cfufw_ignored=0
cfufw_nonew=0
cfufw_purge=0
cfufw_showhelp=0

ufw default allow outgoing
ufw default deny incoming
ufw allow OpenSSH

cf_ufw_add () {
    if [ ! -z $1 ]; then
        rule=$(LC_ALL=C && ufw allow from $1 to any port 80,443 proto tcp comment "cloudflare")

        if [ "$rule" = 'Rule added' ] || [ "$rule" = 'Rule added (v6)' ]; then
            echo -n "\e[32m+\e[39m"
            cfufw_created=$((cfufw_created+1))
            return
        fi
    fi

    echo -n "\e[90m.\e[39m"
    cfufw_ignored=$((cfufw_ignored+1))
}

cf_ufw_del () {
    if [ ! -z $1 ]; then
        rule=$(LC_ALL=C && ufw delete allow from $1 to any port 80,443 proto tcp)

        if [ "$rule" = 'Rule deleted' ] || [ "$rule" = 'Rule deleted (v6)' ]; then
            echo -n "\e[31m-\e[39m"
            cfufw_deleted=$((cfufw_deleted+1))
            return
        fi
    fi

    echo -n "\e[90m.\e[39m"
    cfufw_ignored=$((cfufw_ignored+1))
}

cf_ufw_purge () {
    total="$(ufw status numbered | awk '/# cloudflare$/ {++count} END {print count}')"
    i=1

    if [ -z $total ]; then
        cfufw_deleted=0
        return
    fi

    while [ $i -le $total ]; do
        cfip=$(ufw status numbered | awk '/# cloudflare$/{print $6; exit}')
        cf_ufw_del $cfip
        i=$((i+1))
    done
}

echo '█▀▀ █▀▀   █░█ █▀▀ █░█░█'
echo '█▄▄ █▀░   █▄█ █▀░ ▀▄▀▄▀'
echo ''

for arg in "$@"; do
    case "$arg" in
        '--purge') cfufw_purge=1 ;;
        '-p') cfufw_purge=1 ;;
        '--no-new') cfufw_nonew=1 ;;
        '-n') cfufw_nonew=1 ;;
        '--help') cfufw_showhelp=1 ;;
        '-h') cfufw_showhelp=1 ;;
    esac
done

if [ $cfufw_showhelp -eq 1 ]; then
    echo 'ufw-cf.sh 2.0 (https://github.com/drvy/ufw-cloudflare)'
    echo 'Retrieve Cloudflare IPs and create allow rules in UFW (80 and 443 tcp) for each.'
    echo 'Usage: ./ufw-cf.sh [options]'
    echo 'OPTIONS:'
    echo "\t--help (-h)  : This."
    echo "\t--purge (-p) : Remove existing CF rules (Deletes rules with #cloudflare comment)."
    echo "\t--no-new (-n): Does not download CF IPs and does not add any rule to UFW."
    echo 'EXAMPLES:'
    echo "\t./ufw-cf.sh --purge"
    echo "\t./ufw-cf.sh --purge --no-new"
    exit
fi

if [ $cfufw_purge -eq 1 ]; then
    cf_ufw_purge
fi

if [ $cfufw_nonew -eq 0 ]; then
    [ -e /tmp/cloudflare-ips.txt ] && rm /tmp/cloudflare-ips.txt
    touch /tmp/cloudflare-ips.txt

    wget https://www.cloudflare.com/ips-v4 -q -O ->> /tmp/cloudflare-ips.txt
    echo "" >> /tmp/cloudflare-ips.txt
    wget https://www.cloudflare.com/ips-v6 -q -O ->> /tmp/cloudflare-ips.txt

    for cfip in `cat /tmp/cloudflare-ips.txt`; do
        cf_ufw_add "${cfip}"
    done

    [ -e /tmp/cloudflare-ips.txt ] && rm /tmp/cloudflare-ips.txt
fi

echo ''
echo "Total rules deleted: ${cfufw_deleted}"
echo "Total rules created: ${cfufw_created}"
echo "Total rules ignored: ${cfufw_ignored}"
ufw --force enable
echo 'Done.'

