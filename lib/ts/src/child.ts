import * as net from 'node:net';
import {IPCCommon, type IPCOptions} from './common.js';
import {socketPathFromArgs} from './util.js';

export class ChildIPC extends IPCCommon {
    constructor(opts?: IPCOptions, ...localApis: object[]) {
        super(localApis, socketPathFromArgs(), opts);
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
        const closePromise = new Promise<void>((resolve) => {
            this.onClose = () => {
                if (this.processingCalls === 0) {
                    this.conn?.destroy();
                    resolve();
                }
            };
            if (this.stopRequested && this.processingCalls === 0) {
                this.conn?.destroy();
                resolve();
            }
        });

        const errorPromise = this.errorQueue.collect().then((errors) => {
            if (errors.length === 1) {
                throw errors[0];
            } else if (errors.length > 1) {
                throw new Error(errors.map(e => e.toString()).join(', '));
            }
        });

        await Promise.race([closePromise, errorPromise]);
    }
}
