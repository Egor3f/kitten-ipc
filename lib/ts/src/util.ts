import * as util from 'node:util';

const IPC_SOCKET_ARG = 'ipc-socket';

export function socketPathFromArgs(): string {
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

export function timeout<T>(prom: Promise<T>, ms: number): Promise<T> {
    return Promise.race(
        [
            prom,
            new Promise((res, reject) => {
                setTimeout(() => {reject(new Error('timed out'))}, ms)
            })]
    ) as Promise<T>;
}
