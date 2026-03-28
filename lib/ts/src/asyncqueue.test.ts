import {test} from 'vitest';
import {AsyncQueue} from './asyncqueue.js';

test('put then collect returns items', async ({expect}) => {
    const q = new AsyncQueue<number>();
    q.put(1);
    q.put(2);
    const items = await q.collect();
    expect(items).toEqual([1, 2]);
});

test('collect then put resolves on first put', async ({expect}) => {
    const q = new AsyncQueue<number>();
    const collectPromise = q.collect();
    q.put(42);
    const items = await collectPromise;
    expect(items).toEqual([42]);
});

test('multiple puts before collect', async ({expect}) => {
    const q = new AsyncQueue<string>();
    q.put('a');
    q.put('b');
    q.put('c');
    const items = await q.collect();
    expect(items).toEqual(['a', 'b', 'c']);
});
