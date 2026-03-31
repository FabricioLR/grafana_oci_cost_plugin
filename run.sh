mage -v build:linux
npm run build
docker compose down && docker compose up -d
docker compose logs -f | grep "plugin.transoft-ocicost-datasource"

