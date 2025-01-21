FROM eclipse-temurin:11.0.21_9-jdk-jammy

USER root

# TODO install https://github.com/YS-L/csvlens
RUN apt update && \
    apt install -y tinyproxy dante-server bind9 autossh openssh-server \
                   tcpdump bind9-dnsutils netcat nmap inetutils-ping socat proxychains \
                   sudo kitty apt-file git \
                   htop zsh jq file less \
                   neovim nano \
                   kafkacat \
                   \
                   && \
    apt clean

COPY files /

# Install Trino command line: https://trino.io/docs/current/client/cli.html
RUN cd /tmp && wget https://repo1.maven.org/maven2/io/trino/trino-cli/437/trino-cli-437-executable.jar && \
    mv trino-cli-437-executable.jar trino && chmod +x trino && mv trino /usr/local/bin/


RUN useradd -m tunnel &&  \
    chown -R tunnel:tunnel /home/tunnel/ && \
    mkdir -p home/tunnel/.ssh/ && \
    chmod -R go-rwx /home/tunnel/.ssh/ && \
    usermod tunnel -s /bin/bash && \
    chmod +x /start.sh

CMD /bin/sh -c "/start.sh"
