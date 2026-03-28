import {test} from 'vitest';
import {ParentIPC} from './lib.js';

test('test connection timeout', async ({expect}) => {
    const parentIpc = new ParentIPC('../testdata/sleep15.sh', []);
    await expect(parentIpc.start()).rejects.toThrowError('timed out');
}, 15000);

test('test process stop before connection accept', async ({expect}) => {
    const parentIpc = new ParentIPC('../testdata/sleep3.sh', []);
    await expect(parentIpc.start()).rejects.toThrowError();
}, 15000);
