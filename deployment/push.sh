DOMAIN=$1
docker tag coolcar-"$DOMAIN":latest 402494663939.dkr.ecr.us-west-2.amazonaws.com/coolcar:"$DOMAIN"
docker push 402494663939.dkr.ecr.us-west-2.amazonaws.com/coolcar:"$DOMAIN"
