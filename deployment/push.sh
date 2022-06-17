DOMAIN=$1
TAG=$2
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws/l4j6s3h2
docker tag coolcar-"$DOMAIN":latest public.ecr.aws/l4j6s3h2/coolcar:"$DOMAIN"-"$TAG"
docker push public.ecr.aws/l4j6s3h2/coolcar:"$DOMAIN"-"$TAG"
