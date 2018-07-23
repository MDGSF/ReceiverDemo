const WebSocket = require('ws');
const msgpack = require('msgpack');
const rp = require('request-promise-native');
const WebSocketServer = WebSocket.Server;

const verifyToken = async function (device, token) {
	const url = 'https://setup.minieye.cc/services/report/agent/token/verify';
	const form = {device: device, token: token};
	const options = {method: 'POST', url, form};

	let body, res;

	try {
		body = await rp(options);
	} catch (req) {
		throw new Error(req.message);
	}

	try {
		res = JSON.parse(body);
	} catch (err) {
		throw new Error('malformed response body');
	}

	if (res.result === true) {
		return true;
	} else if (res.error) {
		throw new Error(res.error);
	} else {
		throw new Error('malformed response object');
	}
};

const handleMessage = function (msg) {
	if (!this.established) {
		if (msg.type == 'auth') {

			console.log(JSON.stringify(msg));
			verifyToken(msg.device, msg.token)
			.then(ok => {
				this.sendMessage({type: 'auth_ok', accept: 'all'});
				this.established = true;
				this.device = msg.device;
			})
			.catch(err => {
				this.sendMessage({type: 'auth_error', reason: err.message});
				this.terminate();
			});

		} else {
			this.sendMessage({type: 'error', message: 'unauthorized'});
			this.terminate();
		}
	} else {
		if (msg.type == 'event') {
			console.log(`device[${this.device}], event = ${JSON.stringify(msg)}`);
			this.sendMessage({type: 'ack', id: msg.id, time: Date.now()});
		} else if (msg.type == 'file') {
			console.log(`device[${this.device}], file = ${JSON.stringify(msg)}`);
			this.sendMessage({type: 'ack', id: msg.id, time: Date.now()});
		} else if (msg.type == 'status') {
			console.log(`device[${this.device}], status = ${JSON.stringify(msg)}`);
			this.sendMessage({type: 'ack', id: msg.id, time: Date.now()});
		} else if (msg.type == 'dms') {
			console.log(`device[${this.device}], dms = ${JSON.stringify(msg)}`);
			this.sendMessage({type: 'ack', id: msg.id, time: Date.now()});
		} else {
			console.log(`device[${this.device}], unknown msg ${JSON.stringify(msg)}`);
		}
	}
};

const onMessage = function (data) {
	if (typeof data == 'string') {
		console.log('receive data: %s', data);
	} else {
		const msg = msgpack.unpack(data);
		handleMessage.call(this, msg);
	}
};

const onConnection = function (ws) {
	ws.sendMessage = function (msg) {
		ws.send(msgpack.pack(msg));
	};

	ws.established = false;
	ws.device = null;
	ws.on('message', onMessage.bind(ws));
};

const wss = new WebSocketServer({ host: "127.0.0.1", port: 3000 });
wss.on('connection', onConnection);

