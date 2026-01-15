FROM mirror.gcr.io/library/debian:stable-slim

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y \
    # Base Tools
    micro \
    nano \
    curl \
    wget \
    mc \
    netcat-traditional \
    bash-completion \
    jq \
    htop \
    hstr \
    git \
    # Network Tools
    apache2-utils \
    dnsutils \
    bridge-utils \
    conntrack \
    ethtool \
    file \
    fping \
    iftop \
    iperf \
    iperf3 \
    iproute2 \
    ipset \
    iptables \
    iptraf-ng \
    iputils-ping \
    ipvsadm \
    httpie \
    liboping0 \
    ltrace \
    mtr-tiny \
    snmp \
    nftables \
    ngrep \
    nmap \
    openssl \
    python3-pip \
    python3-setuptools \
    python3-scapy \
    socat \
    speedtest-cli \
    openssh-client \
    strace \
    tcpdump \
    tcptraceroute \
    traceroute \
    tshark \
    ca-certificates \
    libssl3 \
    perl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Setting User and Home
USER root
WORKDIR /root
ENV HOSTNAME=docker-net-tools
ENV SHELL=/bin/bash

# Fix permissions for OpenShift and tshark
RUN chmod -R g=u /root
RUN chown root:root /usr/bin/dumpcap

# Configure bash aliases for root user
RUN cat >> /root/.bashrc <<'EOF'

# Custom configuration
alias ll='ls $LS_OPTIONS -lAFh'

# HSTR configuration
alias hh=hstr                    # hh to be alias for hstr
export HSTR_CONFIG=hicolor       # get more colors
shopt -s histappend              # append new history items to .bash_history
export HISTCONTROL=ignorespace   # leading space hides commands from history
export HISTFILESIZE=10000        # increase history file size (default is 500)
export HISTSIZE=${HISTFILESIZE}  # increase history size (default is 500)
# ensure synchronization between bash memory and history file
export PROMPT_COMMAND="history -a; history -n; ${PROMPT_COMMAND}"
function hstrnotiocsti {
    { READLINE_LINE="$( { </dev/tty hstr ${READLINE_LINE}; } 2>&1 1>&3 3>&- )"; } 3>&1;
    READLINE_POINT=${#READLINE_LINE}
}
# if this is interactive shell, then bind hstr to Ctrl-r (for Vi mode check doc)
if [[ $- =~ .*i.* ]]; then bind -x '"\C-r": "hstrnotiocsti"'; fi
export HSTR_TIOCSTI=n
EOF

# Running BASH
CMD ["/bin/bash"]
