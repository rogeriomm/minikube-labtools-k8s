FROM eclipse-temurin:11.0.21_9-jdk-jammy

USER root

RUN apt update && \
    apt install -y tinyproxy dante-server bind9 autossh openssh-server \
                   tcpdump bind9-dnsutils netcat nmap inetutils-ping \
                   sudo kitty apt-file \
                   htop zsh \
                   neovim nano \
                   kafkacat \
                   && \
    apt clean

COPY files /

RUN useradd -m tunnel &&  \
    chown -R tunnel:tunnel /home/tunnel/ && \
    chmod -R go-rwx /home/tunnel/.ssh/ && \
    usermod tunnel -s /bin/bash && \
    chmod +x /start.sh

CMD /bin/sh -c "/start.sh"
