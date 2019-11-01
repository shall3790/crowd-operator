#!/bin/bash
# set -e
# IMPORTANT: Do not add more content to this file unless you know what you are
#            doing. This file is sourced everytime the shell session is opened.
# This will make scl collection binaries work out of box.
unset BASH_ENV PROMPT_COMMAND ENV
# if head "/etc/redhat-release" | grep -q "^CentOS Linux release 7" || \
#    head "/etc/redhat-release" | grep -q "^Red Hat Enterprise Linux\( Server\)\? release 7"; then
#     source scl_source enable $SCL_PKGS
# fi
# source /opt/app-root/bin/activate
# Set current user in nss_wrapper
USER_ID=$(id -u)
GROUP_ID=$(id -g)
if [ x"$USER_ID" != x"0" -a x"$USER_ID" != x"1001" ]; then
    NSS_WRAPPER_PASSWD=/tmp/passwd
    NSS_WRAPPER_GROUP=/etc/group
    cat /etc/passwd | sed -e 's/^default:/builder:/' > $NSS_WRAPPER_PASSWD
    echo "default:x:${USER_ID}:${GROUP_ID}:buildagent:${HOME}:/sbin/nologin" >> $NSS_WRAPPER_PASSWD
    export NSS_WRAPPER_PASSWD
    export NSS_WRAPPER_GROUP
    LD_PRELOAD=libnss_wrapper.so
    export LD_PRELOAD
fi
exec "$@"