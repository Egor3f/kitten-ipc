import {ChildIPC} from 'kitten-ipc';
import GoIpcApi from './goapi.gen.js';

/**
 * @kittenipc api
 */
class TsIpcApi {
    Div(a: number, b: number): number {
        if (b === 0) {
            throw new Error('division by zero');
        }
        return a / b;
    }
}

async function main() {
    const localApi = new TsIpcApi();
    const ipc = new ChildIPC(localApi);
    const goApi = new GoIpcApi(ipc);

    await ipc.start();

    console.log(`12/3=${await goApi.Div(12, 3)}`);

    try {
        await goApi.Div(10, 0);
    } catch (e) {
        console.trace(e);
    }

    await ipc.wait();
}

main().catch(e => {
    console.trace(e);
});
