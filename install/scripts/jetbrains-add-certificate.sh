#!/usr/bin/env zsh

add_cert()
{
  echo "Jetbrains: $1"
  cd "/Applications/$1/Contents/jbr/Contents/Home/lib/security"

  "/Applications/$1/Contents/jbr/Contents/Home/bin/keytool" -keystore cacerts \
            -importcert -alias minikube-cert -file $MINIKUBE_HOME/ca.crt
}

echo "Adding Minikube certificate on all Jetbrains tools. Enter password 'changeit'"

add_cert "IntelliJ IDEA.app"
add_cert "PyCharm.app"
add_cert "GoLand.app"
add_cert "DataGrip.app"
