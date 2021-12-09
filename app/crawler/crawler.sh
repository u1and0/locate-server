#!/bin/bash
# Oneliner
#
# ```sh
# curl -fsSL "localhost:8080/history?gt=100" |
#   jq '.[] |
#   .word' |
#   tr -d \" |
#   nkf -WwMQ |
#   sed -e 's/=$//g' |
#   tr = % |
#   xargs -I@ curl -fsSL "localhost:8080/json?q=@&logging=false"
# ```
#
# function urlencoding {
#     echo "$*" | nkf -WwMQ | sed -e 's/=$//g' | tr = % | tr -d '\n'
# }

url="localhost:8080"
# his=$(curl -fsSL "${url}/history?gt=100" | jq '.[] | .word')
# echo ${his} | xargs urlencoding | xargs -I@ curl "${url}/json?q=@&logging=false"
curl -fsSL "${url}/history?gt=1" |  # fetch history
  jq '.[] | .word' | # flatten list
  tr -d \" |  # trim double quote
  nkf -WwMQ |  # URL encoding
  sed -e 's/=$//g' |  # URL encoding
  tr = % |  # URL encoding
  xargs -I@ curl -fsSL "${url}/json?q=@&logging=false" > /dev/null 2>&1 # cache json & trash output
