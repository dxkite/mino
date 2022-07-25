import { defineComponent } from "vue";


const webSocketMixin = defineComponent<any,any,{
    wsLink: string;
    wsType: string;
    socket: WebSocket | undefined;
}>({
    data() {
        return {
            wsLink: "",
            wsType: 'json',
            socket: undefined,
        };
    },
    mounted() {
        this._ws$init();
    },
    methods: {
        _ws$init() {
            // console.log('Init Websocket');
            this.socket = new WebSocket(this.wsLink)
            this.socket.addEventListener('open', () => {
                // console.log('open');
                this._ws$open();
            });
            this.socket.addEventListener('close', (evt: any) => {
                console.log('close', evt);
                this._ws$close();
            });
            this.socket.addEventListener('error', (evt: any) => {
                console.log('error', evt);
                this._ws$error();
            });
            this.socket.addEventListener('message', (evt: {data: string}) => {
                this._ws$message(evt.data);
            });
        },
        _ws$open() {
            // console.log('open');
            if (this.onWsOpen) {
                this.onWsOpen();
            }
        },
        _ws$message(msg: string) {
            if (this.onWsMessage) {
                if (this.wsType === 'json') {
                    msg = JSON.parse(msg);
                }
                this.onWsMessage(msg);
            }
        },
        $send(params: string) {
            this.socket.send(params)
        },
        _ws$close() {
            if (this.onWsClose) {
                this.onWsClose();
            }
        },
        _ws$error() {
            if (this.onWsError) {
                this.onWsError();
            } else {
                this.$notify({
                    type: "error",
                    title: "链接websocket失败",
                    message: this.wsLink,
                });
            }
        },
    }
});

export default webSocketMixin;