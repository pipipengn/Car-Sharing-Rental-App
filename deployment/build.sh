DOMAIN=$1
# shellcheck disable=SC2164
cd ../server
docker build -t coolcar-"$DOMAIN" -f ../deployment/"$DOMAIN"/Dockerfile .


