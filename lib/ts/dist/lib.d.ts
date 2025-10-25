import * as net from 'node:net';
import { QueuedEvent } from 'ts-events';
declare enum MsgType {
    Call = 1,
    Response = 2
}
type Vals = any[];
interface CallMessage {
    type: MsgType.Call;
    id: number;
    method: string;
    params: Vals;
}
interface ResponseMessage {
    type: MsgType.Response;
    id: number;
    result?: Vals;
    error?: string;
}
type Message = CallMessage | ResponseMessage;
interface CallResult {
    result: Vals;
    error: Error | null;
}
declare abstract class IPCCommon {
    protected localApi: any;
    protected socketPath: string;
    protected conn: net.Socket | null;
    protected nextId: number;
    protected pendingCalls: Record<number, (result: CallResult) => void>;
    protected errors: QueuedEvent<Error>;
    protected constructor(localApi: any, socketPath: string);
    protected readConn(): void;
    protected processMsg(msg: Message): void;
    protected handleCall(msg: CallMessage): void;
    protected sendMsg(msg: Message): void;
    protected handleResponse(msg: ResponseMessage): void;
    protected raiseErr(err: Error): void;
    call(method: string, ...params: Vals): Promise<Vals>;
}
export declare class ParentIPC extends IPCCommon {
    private readonly cmdPath;
    private readonly cmdArgs;
    private cmd;
    private readonly listener;
    constructor(cmdPath: string, cmdArgs: string[], localApi: any);
    start(): Promise<void>;
    private acceptConn;
    wait(): Promise<void>;
}
export declare class ChildIPC extends IPCCommon {
    constructor(localApi: any);
    start(): Promise<void>;
    wait(): Promise<void>;
}
export {};
//# sourceMappingURL=lib.d.ts.map