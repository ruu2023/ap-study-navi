#!/usr/bin/env bash
# Build the image locally for the server's arch, ship it over SSH, and restart.
# Usage: DEPLOY_HOST=ubuntu@13.159.19.55 server/deploy/deploy.sh
set -euo pipefail
HOST=${DEPLOY_HOST:-ubuntu@13.159.19.55}
DIR=/opt/ap-study
cd "$(dirname "$0")/.."

TAG=$(date +%Y%m%d%H%M%S)
echo "==> build ap-study:$TAG (linux/amd64)"
docker buildx build --platform linux/amd64 --load -t "ap-study:$TAG" -t ap-study:latest .

echo "==> ship image"
docker save "ap-study:$TAG" | gzip | ssh "$HOST" 'gunzip | docker load'
# Also tag :latest so a plain `docker compose up -d` on the server works.
ssh "$HOST" "docker tag ap-study:$TAG ap-study:latest"

echo "==> sync compose and secrets"
ssh "$HOST" "sudo mkdir -p $DIR && sudo chown \$(id -u):\$(id -g) $DIR"
rsync -az deploy/compose.yaml "$HOST:$DIR/compose.yaml"
# Secrets are generated on the server on first deploy and never leave it.
ssh "$HOST" "cd $DIR && if [ ! -f .env ]; then umask 077; printf 'AP_PASSWORD=%s\nAP_SECRET=%s\n' \"\$(openssl rand -base64 18 | tr -d '/+=' | cut -c1-16)\" \"\$(openssl rand -hex 32)\" > .env; echo 'generated .env'; fi"

echo "==> restart"
ssh "$HOST" "cd $DIR && APP_TAG=$TAG docker compose up -d"
# Keep only the current and previous ap-study images (other projects untouched).
ssh "$HOST" "docker images ap-study --format '{{.Tag}}' | grep -v latest | sort -r | tail -n +3 | xargs -r -I{} docker rmi ap-study:{} >/dev/null || true"
ssh "$HOST" "for i in \$(seq 1 20); do docker exec lab-ruu-caddy-1 wget -qO- http://ap-study:8080/ap/healthz 2>/dev/null && echo ' healthy' && exit 0; sleep 1; done; echo 'not healthy'; docker logs --tail 30 ap-study; exit 1"
