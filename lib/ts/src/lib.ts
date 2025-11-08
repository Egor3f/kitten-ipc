import * as net from 'node:net';
import * as readline from 'node:readline';
import {type ChildProcess, spawn} from 'node:child_process';
import * as os from 'node:os';
import * as path from 'node:path';
import * as fs from 'node:fs';
import * as util from 'node:util';
import {AsyncQueue} from './asyncqueue.js';

const IPC_SOCKET_ARG = 'ipc-socket';

enum MsgType {
    Call = 1,
    Response = 2,
}

type Vals = any[];

interface CallMessage {
    type: MsgType.Call,
    id: number,
    method: string;
    params: Vals;
}

interface ResponseMessage {
    type: MsgType.Response,
    id: number,
    result?: Vals;
    error?: string;
}

type Message = CallMessage | ResponseMessage;

interface CallResult {
    result: Vals;
    error: Error | null;
}

abstract class IPCCommon {
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
                this.handleCall(msg).catch(this.errorQueue.put);
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
        const method = endpoint[methodName];
        if (!method || typeof method !== 'function') {
            this.sendMsg({type: MsgType.Response, id: msg.id, error: `method not found: ${ msg.method }`});
            return;
        }

        const argsCount = method.length;
        if (msg.params.length !== argsCount) {
            this.sendMsg({
                type: MsgType.Response,
                id: msg.id,
                error: `argument count mismatch: expected ${ argsCount }, got ${ msg.params.length }`
            });
            return;
        }

        try {
            this.processingCalls++;
            let result = method.apply(endpoint, msg.params);
            if (result instanceof Promise) {
                result = await result;
            }
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

    call(method: string, ...params: Vals): Promise<Vals> {
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
                this.sendMsg({type: MsgType.Call, id, method, params});
            } catch (e) {
                delete this.pendingCalls[id];
                reject(new Error(`send call: ${ e }`));
            }
        });
    }

    protected raiseErr(err: Error): void {
        this.errorQueue.put(err);
    }
}


export class ParentIPC extends IPCCommon {
    private readonly cmdPath: string;
    private readonly cmdArgs: string[];
    private cmd: ChildProcess | null = null;
    private readonly listener: net.Server;

    constructor(cmdPath: string, cmdArgs: string[], ...localApis: object[]) {
        const socketPath = path.join(os.tmpdir(), `kitten-ipc-${ process.pid }.sock`);
        super(localApis, socketPath);

        this.cmdPath = cmdPath;
        if (cmdArgs.includes(`--${ IPC_SOCKET_ARG }`)) {
            throw new Error(`you should not use '--${ IPC_SOCKET_ARG }' argument in your command`);
        }
        this.cmdArgs = cmdArgs;

        this.listener = net.createServer();
    }

    async start(): Promise<void> {
        try {
            fs.unlinkSync(this.socketPath);
        } catch {
        }

        await new Promise<void>((resolve, reject) => {
            this.listener.listen(this.socketPath, () => {
                resolve();
            });
            this.listener.on('error', reject);
        });

        const cmdArgs = [...this.cmdArgs, `--${ IPC_SOCKET_ARG }`, this.socketPath];
        this.cmd = spawn(this.cmdPath, cmdArgs, {stdio: 'inherit'});

        this.cmd.on('error', (err) => {
            this.raiseErr(err);
        });

        this.acceptConn().catch();
    }

    private async acceptConn(): Promise<void> {
        const acceptTimeout = 10000;

        const acceptPromise = new Promise<net.Socket>((resolve, reject) => {
            this.listener.once('connection', (conn) => {
                resolve(conn);
            });
            this.listener.once('error', reject);
        });

        try {
            this.conn = await timeout(acceptPromise, acceptTimeout);
            this.readConn();
        } catch (e) {
            if (this.cmd) this.cmd.kill();
            this.raiseErr(e as Error);
        }
    }

    async wait(): Promise<void> {
        return new Promise(async (resolve, reject) => {
            if (!this.cmd) {
                reject('Command is not started yet');
                return;
            }
            this.cmd.addListener('close', (code, signal) => {
                if (signal || code) {
                    if (signal) reject(new Error(`Process exited with signal ${ signal }`));
                    else reject(new Error(`Process exited with code ${ code }`));
                } else if(!this.ready) {
                    reject('command exited before connection established');
                } else {
                    resolve();
                }
            });
            const errors = await this.errorQueue.collect();
            if(errors.length === 1) {
                reject(errors[0]);
            } else if(errors.length > 1) {
                reject(new Error(errors.map(Error.toString).join(', ')));
            }
        });
    }
}


export class ChildIPC extends IPCCommon {
    constructor(...localApis: object[]) {
        super(localApis, socketPathFromArgs());
    }

    async start(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.conn = net.createConnection(this.socketPath, () => {
                this.readConn();
                resolve();
            });
            this.conn.on('error', reject);
        });
    }

    async wait(): Promise<void> {
        return new Promise(async (resolve, reject) => {
            const errors = await this.errorQueue.collect();
            if(errors.length === 1) {
                reject(errors[0]);
            } else if(errors.length > 1) {
                reject(new Error(errors.map(Error.toString).join(', ')));
            }
            this.onClose = () => {
                if (this.processingCalls === 0) {
                    this.conn?.destroy();
                    resolve();
                }
            };
        });
    }
}


function socketPathFromArgs(): string {
    const {values} = util.parseArgs({
        options: {
            [IPC_SOCKET_ARG]: {
                type: 'string',
            }
        }
    });

    if (!values[IPC_SOCKET_ARG]) {
        throw new Error('ipc socket path is missing');
    }

    return values[IPC_SOCKET_ARG];
}


function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}


// throws on timeout
function timeout<T>(prom: Promise<T>, ms: number): Promise<T> {
    return Promise.race(
        [
            prom,
            new Promise((res, reject) => {
                setTimeout(() => {reject(new Error('timed out'))}, ms)
            })]
    ) as Promise<T>;
}
