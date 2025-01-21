# openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout privateKey.key -out certificate.crt
openssl x509 -inform pem -in certificate.crt -out world.xpt.pem