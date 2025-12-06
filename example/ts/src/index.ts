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
    XorData(data1: Buffer, data2: Buffer): Buffer {
        if (data1.length === 0 || data2.length === 0) {
            throw new Error('empty input data');
        }
        if (data1.length !== data2.length) {
            throw new Error('input data length mismatch');
        }
        const result = Buffer.alloc(data1.length);
        for (let i = 0; i < data1.length; i++) {
            result[i] = data1[i]! ^ data2[i]!;
        }
        return result;
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
