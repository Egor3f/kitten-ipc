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

    console.log(`call result ts->go Div = ${await remoteApi.Div(10, 2)}`);

    const data1 = Buffer.alloc(10, 0b10101010);
    const data2 = Buffer.alloc(10, 0b11110000);
    console.log(`call result ts->go XorData = ${(await remoteApi.XorData(data1, data2)).toString('hex')}`);

    await ipc.wait();
}

main().catch(e => {
    console.trace(e);
});
