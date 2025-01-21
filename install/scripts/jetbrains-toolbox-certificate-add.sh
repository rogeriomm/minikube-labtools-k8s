#!/usr/bin/env zsh

TOOLBOX_HOME="$HOME/Library/Application Support/JetBrains/Toolbox/apps"

add_cert()
{
  echo "Jetbrains: $1="
  APP_PATH="$TOOLBOX_HOME/$2/$1"
  cd "$APP_PATH/Contents/jbr/Contents/Home/lib/security" || exit

  "$APP_PATH/Contents/jbr/Contents/Home/bin/keytool" -keystore cacerts \
            -importcert -alias minikube-cert -file "$MINIKUBE_HOME"/ca.crt
}

echo "Adding Minikube certificate on all Jetbrains tools. Enter password 'changeit'"

add_cert "IntelliJ IDEA.app" "IDEA-U/ch-0/213.6461.79"
add_cert "PyCharm.app"       "PyCharm-P/ch-0/213.6461.77"
add_cert "GoLand.app"        "Goland/ch-0/213.6461.81"
add_cert "DataGrip.app"      "datagrip/ch-0/213.6461.82/"
