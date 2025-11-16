import {ChildIPC} from 'kitten-ipc';
import GoIpcApi from './remote.js';

/**
 * @kittenipc api
 */
class TsIpcApi {
    Div(a: number, b: number): number {
        if (b === 0) {
            throw new Error('zero division');
        }
        return a / b;
    }
}

async function main() {
    const localApi = new TsIpcApi();
    const ipc = new ChildIPC(localApi);
    const remoteApi = new GoIpcApi(ipc);

    await ipc.start();

    console.log(`remote call result from go = ${await remoteApi.Div(10, 2)}`);

    await ipc.wait();
}

main().catch(e => {
    console.trace(e);
});
