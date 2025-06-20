{pkgs, ...}: {
  env.GREET = "devenv";
  dotenv.enable = true;

  packages = [pkgs.git];

  scripts.deploy-discord.exec = ''
    cd $DEVENV_ROOT/discord-bot

    echo "Building discord-bot..."
    docker buildx build --platform linux/amd64 -t $SCW_CR_NAME/pz-discord-bot:latest --push .

    echo "Done"
  '';

  scripts.deploy-srv-mon.exec = ''
    echo "Checking server..."
    if ! ping -c 1 -W 5 "$INSTANCE_IP" > /dev/null 2>&1; then
        echo "Error: Host $INSTANCE_IP is not reachable"
        exit 1
    fi

    cd $DEVENV_ROOT/server-monitor

    echo "building srv-mon..."
    GOOS=linux GOARCH=amd64 go build -o srv-mon

    echo "pushing srv-mon to instance..."
    scp srv-mon $INSTANCE_DEST/srv-mon

    echo "pushing .env to instance..."
    scp $DEVENV_ROOT/.env $INSTANCE_DEST/.env

    echo "Done"
  '';

  scripts.deploy-listener-func.exec = ''
    cd $DEVENV_ROOT/listner-func/

    echo "Deploying listener..."
    serverless deploy
  '';
}
