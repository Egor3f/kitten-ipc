import {test} from 'vitest';
import {ParentIPC} from './lib.js';

test('test connection timeout', async ({expect}) => {
    const parentIpc = new ParentIPC('testdata/sleep15.sh', []);
    await parentIpc.start();
    await expect(parentIpc.wait()).rejects.toThrowError('timed out');
});

test('test process stop before connection accept', async ({expect}) => {
    const parentIpc = new ParentIPC('testdata/sleep3.sh', []);
    await parentIpc.start();
    await expect(parentIpc.wait()).rejects.toThrowError('command exited before connection established');
});
