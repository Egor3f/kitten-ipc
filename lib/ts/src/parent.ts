import * as net from 'node:net';
import * as os from 'node:os';
import * as path from 'node:path';
import * as fs from 'node:fs';
import * as crypto from 'node:crypto';
import {type ChildProcess, spawn} from 'node:child_process';
import {IPCCommon, type IPCOptions} from './common.js';
import {timeout} from './util.js';

const IPC_SOCKET_ARG = 'ipc-socket';
const ACCEPT_TIMEOUT_MS = 10000;

export class ParentIPC extends IPCCommon {
    private readonly cmdPath: string;
    private readonly cmdArgs: string[];
    private cmd: ChildProcess | null = null;
    private readonly listener: net.Server;
    private cmdExitResult: { code: number | null, signal: string | null } | null = null;
    private cmdExitCallbacks: ((result: { code: number | null, signal: string | null }) => void)[] = [];

    constructor(cmdPath: string, cmdArgs: string[], opts?: IPCOptions, ...localApis: object[]) {
        const socketPath = path.join(os.tmpdir(), `kitten-ipc-${ process.pid }-${ crypto.randomInt(2**48 - 1) }.sock`);
        super(localApis, socketPath, opts);

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

        this.cmd.on('close', (code, signal) => {
            const result = { code, signal };
            this.cmdExitResult = result;
            for (const cb of this.cmdExitCallbacks) cb(result);
            this.cmdExitCallbacks = [];
        });

        await this.acceptConn();
    }

    private async acceptConn(): Promise<void> {
        const acceptPromise = new Promise<net.Socket>((resolve, reject) => {
            this.listener.once('connection', (conn) => {
                resolve(conn);
            });
            this.listener.once('error', reject);
        });

        const exitPromise = new Promise<net.Socket>((_, reject) => {
            if (this.cmdExitResult) {
                reject(new Error(`command exited before connection established`));
            } else {
                this.cmdExitCallbacks.push(() => {
                    reject(new Error(`command exited before connection established`));
                });
            }
        });

        try {
            this.conn = await timeout(Promise.race([acceptPromise, exitPromise]), ACCEPT_TIMEOUT_MS);
            this.readConn();
        } catch (e) {
            if (this.cmd) this.cmd.kill();
            throw e;
        }
    }

    async wait(): Promise<void> {
        if (!this.cmd) {
            throw new Error('Command is not started yet');
        }

        const exitPromise = new Promise<{ code: number | null, signal: string | null }>((resolve) => {
            if (this.cmdExitResult) {
                resolve(this.cmdExitResult);
            } else {
                this.cmdExitCallbacks.push(resolve);
            }
        });

        try {
            await Promise.race([
                exitPromise.then(({ code, signal }) => {
                    if (signal || code) {
                        if (signal) throw new Error(`Process exited with signal ${ signal }`);
                        else throw new Error(`Process exited with code ${ code }`);
                    } else if (!this.ready) {
                        throw new Error('command exited before connection established');
                    }
                }),
                this.errorQueue.collect().then((errors) => {
                    if (errors.length === 1) {
                        throw errors[0];
                    } else if (errors.length > 1) {
                        throw new Error(errors.map(e => e.toString()).join(', '));
                    }
                }),
            ]);
        } finally {
            try { fs.unlinkSync(this.socketPath); } catch {}
        }
    }
}
