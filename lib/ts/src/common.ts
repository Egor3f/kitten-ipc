import * as net from 'node:net';
import * as readline from 'node:readline';
import {AsyncQueue} from './asyncqueue.js';
import type {CallMessage, CallResult, Message, ResponseMessage, Vals} from './protocol.js';
import {MsgType} from './protocol.js';

export abstract class IPCCommon {
    protected localApis: Record<string, any>;
    protected socketPath: string;
    protected conn: net.Socket | null = null;
    protected nextId: number = 0;
    protected pendingCalls: Record<number, (result: CallResult) => void> = {};
    protected stopRequested: boolean = false;
    protected processingCalls: number = 0;
    protected ready = false;

    protected errorQueue = new AsyncQueue<Error>();
    protected onClose?: () => void;

    protected constructor(localApis: object[], socketPath: string) {
        this.socketPath = socketPath;

        this.localApis = {};
        for (const localApi of localApis) {
            this.localApis[localApi.constructor.name] = localApi;
        }
    }

    protected readConn(): void {
        if (!this.conn) throw new Error('no connection');

        const rl = readline.createInterface({
            input: this.conn,
            crlfDelay: Infinity,
        });

        this.conn.on('error', (e) => {
            this.raiseErr(e);
        });

        this.conn.on('close', (hadError: boolean) => {
            this.rejectPendingCalls(new Error('connection closed'));
            if (hadError) {
                this.raiseErr(new Error('connection closed due to error'));
            }
        });

        rl.on('line', (line) => {
            try {
                const msg: Message = JSON.parse(line);
                this.processMsg(msg);
            } catch (e) {
                this.raiseErr(new Error(`${ e }`));
            }
        });

        this.ready = true;
    }

    protected processMsg(msg: Message): void {
        switch (msg.type) {
            case MsgType.Call:
                this.handleCall(msg).catch((e) => this.errorQueue.put(e));
                break;
            case MsgType.Response:
                this.handleResponse(msg);
                break;
        }
    }

    protected sendMsg(msg: Message): void {
        if (!this.conn) throw new Error('no connection');

        try {
            const data = JSON.stringify(msg) + '\n';
            this.conn.write(data);
        } catch (e) {
            this.raiseErr(new Error(`send response for ${ msg.id }: ${ e }`));
        }
    }

    protected async handleCall(msg: CallMessage) {
        const [endpointName, methodName] = msg.method.split('.');
        if (!endpointName || !methodName) {
            this.sendMsg({type: MsgType.Response, id: msg.id, error: `call malformed: ${ msg.method }`});
            return;
        }
        const endpoint = this.localApis[endpointName];
        if (!endpoint) {
            this.sendMsg({type: MsgType.Response, id: msg.id, error: `endpoint not found: ${ endpointName }`});
            return;
        }
        const method: Function = endpoint[methodName];
        if (!method || typeof method !== 'function') {
            this.sendMsg({type: MsgType.Response, id: msg.id, error: `method not found: ${ msg.method }`});
            return;
        }

        const argsCount = method.length;
        if (msg.args.length !== argsCount) {
            this.sendMsg({
                type: MsgType.Response,
                id: msg.id,
                error: `argument count mismatch: expected ${ argsCount }, got ${ msg.args.length }`
            });
            return;
        }

        try {
            this.processingCalls++;
            let result = method.apply(endpoint, msg.args.map(this.deserialize));
            if (result instanceof Promise) {
                result = await result;
            }
            result = this.serialize(result);
            this.sendMsg({type: MsgType.Response, id: msg.id, result: [result]});
        } catch (err) {
            this.sendMsg({type: MsgType.Response, id: msg.id, error: `${ err }`});
        } finally {
            this.processingCalls--;
        }

        if (this.stopRequested) {
            if (this.onClose) this.onClose();
        }
    }

    protected handleResponse(msg: ResponseMessage): void {
        const callback = this.pendingCalls[msg.id];
        if (!callback) {
            this.raiseErr(new Error(`received response for unknown msgId: ${ msg.id }`));
            return;
        }

        delete this.pendingCalls[msg.id];

        const err = msg.error ? new Error(`remote error: ${ msg.error }`) : null;
        callback({result: msg.result || [], error: err});
    }

    call(method: string, ...args: Vals): Promise<Vals> {
        return new Promise((resolve, reject) => {
            const id = this.nextId++;

            this.pendingCalls[id] = (result: CallResult) => {
                if (result.error) {
                    reject(result.error);
                } else {
                    resolve(result.result);
                }
            };
            try {
                this.sendMsg({type: MsgType.Call, id, method, args: args.map(this.serialize)});
            } catch (e) {
                delete this.pendingCalls[id];
                reject(new Error(`send call: ${ e }`));
            }
        });
    }

    public serialize(arg: any): any {
        switch (typeof arg) {
            case 'string':
            case 'boolean':
            case 'number':
                return arg;
            case 'object':
                if(arg instanceof Buffer) {
                    return arg.toString('base64');
                } else {
                    throw new Error(`cannot serialize ${arg}`);
                }
            default:
                throw new Error(`cannot serialize ${typeof arg}`);
        }
    }

    public deserialize(arg: any): any {
        switch (typeof arg) {
            case 'string':
            case 'boolean':
            case 'number':
                return arg;
            case 'object':
                const keys = Object.entries(arg).map(p => p[0]).sort();
                if(keys[0] === 'd' && keys[1] === 't') {
                    const type = arg['t'];
                    const data = arg['d'];
                    switch (type) {
                        case 'blob':
                            return Buffer.from(data, 'base64');
                        default:
                            throw new Error(`custom object type ${type} is not supported`);
                    }
                } else {
                    throw new Error(`cannot deserialize object with keys ${keys}`);
                }
            default:
                throw new Error(`cannot deserialize ${typeof arg}`);
        }
    }

    stop() {
        if (this.stopRequested) {
            throw new Error('close already requested');
        }
        if (!this.conn || this.conn.readyState === 'closed') {
            throw new Error('connection already closed');
        }
        this.stopRequested = true;
        if (this.onClose) this.onClose();
    }

    protected rejectPendingCalls(err: Error): void {
        const pending = this.pendingCalls;
        this.pendingCalls = {};
        for (const callback of Object.values(pending)) {
            callback({result: [], error: err});
        }
    }

    protected raiseErr(err: Error): void {
        this.errorQueue.put(err);
    }
}
