docker build -t minisrv .
docker rmi localhost:6000/minisrv
docker tag minisrv localhost:6000/minisrv
docker push localhost:6000/minisrv
