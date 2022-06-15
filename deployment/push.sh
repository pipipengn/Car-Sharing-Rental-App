DOMAIN=$1
Account=$2
docker tag coolcar-"$DOMAIN":latest "$Account".dkr.ecr.us-west-2.amazonaws.com/coolcar:"$DOMAIN"
docker push "$Account".dkr.ecr.us-west-2.amazonaws.com/coolcar:"$DOMAIN"
