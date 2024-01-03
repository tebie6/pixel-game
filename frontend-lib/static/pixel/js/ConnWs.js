function ConnWs(serverHost, actionCallback, uuidCallback, openCallback, closeCallback, userToken, loadMode) {
    this.serverHost = serverHost;
    this.actionCallback = actionCallback;
    this.openCallback = openCallback;
    this.closeCallback = closeCallback;
    this.uuidCallback = uuidCallback;
    this.userToken = userToken;
    this.loadMode = loadMode;

    this.wsconnect(); // go
}
// var totalDataSize = 0

ConnWs.prototype = {
    socket: null,	// 连接句柄
    hr_connect_time: 0, // 连接时间
    hb_ping: 0,		// 上次ping时间（戳，毫秒）
    hb_interval: 20000,		// 心跳间隔（毫秒）
    hb_overtime: 30000,		// 重连间隔
    hb_max_overtime: 300000,// 最大重连间隔
    /**
     * 执行连接并开启心跳
     */
    wsconnect: function () {
        var that = this;
        that.initSocket();

        setInterval(function () {
            that.heartbeat();
        }, that.hb_interval);
    },

    /**
     * 重连
     */
    wsreconnect: function () {
        this.socket.close();
        this.initSocket();
        // 重置心跳时间
        this.hb_ping = (new Date).getTime();
    },
    closeConnect: function () {
        this.sendMSG("auth", {
            token: ""
        });
        this.socket.close();
    },
    /**
     * 初始化 socket
     * @param uuid
     */
    initSocket: function (uuid) {

        var that = this;
        this.socket = new WebSocket(`${this.serverHost}/connect`);
        this.socket.binaryType = "arraybuffer";

        // Message received on the socket
        this.socket.onmessage = function (event) {

            // let msg = JSON.parse(event.data);
            // let msgid = msg.action;
            //
            // console.log("Received Message: ", msgid, msg);
            //
            // switch (msgid) {
            //     case 'pong': // pong
            //         that.updateHeartbeat();
            //         break;
            //     default:
            //         that.actionCallback(msgid, msg);
            //         break;
            // }

            if (event.data instanceof ArrayBuffer) {
                try {
                    // 将 Arraybuffer 转换为 Uint8Array
                    const data = new Uint8Array(event.data);

                    // 使用 pako 解压
                    const decompressed = pako.inflate(data);

                    // 解析内容
                    const textDecoder = new TextDecoder('utf-8');
                    const jsonString = textDecoder.decode(decompressed);

                    // 解码包体
                    let msg = JSON.parse(jsonString);
                    let msgid = msg.action;

                    console.log( "Received Message: " , msg);

                    // 只要有响应就认为可用，用来应对重试风暴
                    that.updateHeartbeat();

                    switch (msgid) {
                        case 'pong': // pong
                            // that.updateHeartbeat();
                            break;
                        default:
                            that.actionCallback(msgid, msg);
                            break;
                    }

                } catch (err) {
                    console.error('解析错误:', err);
                }
            }
        };

        this.socket.onopen = function () {
            console.log('that.userToken', )
            // Web Socket 已连接上，使用 send() 方法发送数据
            that.hr_connect_time = that.hb_ping = (new Date).getTime();
            console.log("Ws has opened, update HB ", that.hb_ping);
            that.uuidCallback("");
            that.sendMSG("auth", {
                token: that.userToken,
                size: parseInt(localStorage.getItem("size")),
                load_mode: that.loadMode
            });
            that.openCallback();
        };

        this.socket.onclose = function () {
            // 关闭 websocket
            console.log("Ws has closed");
            that.closeCallback();
        };

        this.socket.onerror = function (event) {
            console.log("Ws error : ", event.data);
        };
    },

    /**
     * 收到服务器 pong 消息
     */
    updateHeartbeat: function () {
        this.hb_ping = (new Date).getTime();
        console.log("revived pong, update HB ", this.hb_ping);
    },

    /**
     * 心跳
     */
    heartbeat: function () {
        // 是否超时重连
        var now = (new Date).getTime();
        var cha = now - this.hb_ping;
        if (cha >= this.hb_overtime) {
            // 如果当前的超时时间小于最大超时时间，则增加
            if (this.hb_overtime < this.hb_max_overtime) {
                this.hb_overtime = Math.min(this.hb_overtime * 2, this.hb_max_overtime);
            }
            this.wsreconnect();
        } else {
            var msgid = "ping";
            var content = {
                timestamp: parseInt(now / 1000),
            };
            this.sendMSG(msgid, content);
        }

    },

    /**
     * 发送消息
     * @param msgid
     * @param msgBody
     */
    sendMSG: function (msgid, msgBody) {
        msgBody.action = msgid;
        console.log(msgBody)
        // 发送
        this.socket.send(JSON.stringify(msgBody));
    }
};










