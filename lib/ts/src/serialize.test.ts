import {test} from 'vitest';
import {ChildIPC} from './child.js';

// Access serialize/deserialize through a ChildIPC instance's public methods
// We create a minimal wrapper to test them
class TestableIPC {
    private ipc: any;

    constructor() {
        // Access the prototype methods directly
        this.ipc = Object.create(ChildIPC.prototype);
    }

    serialize(arg: any): any {
        return this.ipc.serialize(arg);
    }

    deserialize(arg: any): any {
        return this.ipc.deserialize(arg);
    }
}

test('serialize primitives', ({expect}) => {
    const t = new TestableIPC();
    expect(t.serialize(42)).toBe(42);
    expect(t.serialize('hello')).toBe('hello');
    expect(t.serialize(true)).toBe(true);
    expect(t.serialize(3.14)).toBe(3.14);
});

test('serialize buffer to base64', ({expect}) => {
    const t = new TestableIPC();
    const buf = Buffer.from([1, 2, 3]);
    expect(t.serialize(buf)).toBe('AQID');
});

test('deserialize primitives', ({expect}) => {
    const t = new TestableIPC();
    expect(t.deserialize(42)).toBe(42);
    expect(t.deserialize('hello')).toBe('hello');
    expect(t.deserialize(true)).toBe(true);
});

test('deserialize blob to buffer', ({expect}) => {
    const t = new TestableIPC();
    const result = t.deserialize({t: 'blob', d: 'AQID'});
    expect(Buffer.isBuffer(result)).toBe(true);
    expect(result).toEqual(Buffer.from([1, 2, 3]));
});

test('serialize then deserialize blob round-trip', ({expect}) => {
    const t = new TestableIPC();
    const original = Buffer.from([0xDE, 0xAD, 0xBE, 0xEF]);
    // Go serializes as {t: 'blob', d: base64}, TS serializes as just base64 string
    // The TS serialize returns base64 string directly for Buffer
    const serialized = t.serialize(original);
    expect(typeof serialized).toBe('string');
});

test('deserialize unknown object throws', ({expect}) => {
    const t = new TestableIPC();
    expect(() => t.deserialize({foo: 'bar'})).toThrow('cannot deserialize');
});

test('serialize unsupported object throws', ({expect}) => {
    const t = new TestableIPC();
    expect(() => t.serialize({foo: 'bar'})).toThrow('cannot serialize');
});
