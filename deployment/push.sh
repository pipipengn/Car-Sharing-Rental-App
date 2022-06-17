DOMAIN=$1
TAG=$2
PPP=$3
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws/"$PPP"
docker tag coolcar-"$DOMAIN":latest public.ecr.aws/"$PPP"/coolcar:"$DOMAIN"-"$TAG"
docker push public.ecr.aws/"$PPP"/coolcar:"$DOMAIN"-"$TAG"
