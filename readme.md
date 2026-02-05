# 本地linux服务器go使用国内源

    cat <<EOF >> ~/.bashrc
    export GOPROXY=https://goproxy.cn,direct
    export GONOSUMDB=*
    EOF
    source ~/.bashrc

# 安装swagger命令行

    go install github.com/swaggo/swag/cmd/swag@latest

# linux 下安装完swagger需要设置环境变量

    echo 'export PATH=$PATH:/root/go/bin' >> /root/.bashrc
    source /root/.bashrc

# 初始化go项目

    go mod vendor

# 初始化swagger文档

    swag init

# nginx示例配置
        
    upstream bgu {
        server 127.0.0.1:9001;
    }
    server {
        listen 80;
        server_name test.demo.com _;
        location / {
            proxy_pass http://bgu;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_cache_bypass $http_upgrade;
            add_header Access-Control-Allow-Methods *;
            add_header Access-Control-Allow-Origin $http_origin;
        }
    }

# docker示例命令

    docker rm -f bgu-1 && docker rmi bgu-1 &&  docker build --build-arg BUILDKIT_INLINE_CACHE=1 --memory 1GB -t bgu-1 . && docker run -e TZ=Asia/Shanghai -p 9001:8080 --name bgu-1 --restart always -d bgu-1 &&docker logs -t -f bgu-1


go test -v ./core/services -run TestGenerateThunderIndexes