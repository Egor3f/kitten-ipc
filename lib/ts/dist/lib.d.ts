import * as net from 'node:net';
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
    protected localApis: Record<string, any>;
    protected socketPath: string;
    protected conn: net.Socket | null;
    protected nextId: number;
    protected pendingCalls: Record<number, (result: CallResult) => void>;
    protected closeRequested: boolean;
    protected processingCalls: number;
    protected onError?: (err: Error) => void;
    protected onClose?: () => void;
    protected constructor(localApis: object[], socketPath: string);
    protected readConn(): void;
    protected processMsg(msg: Message): void;
    protected sendMsg(msg: Message): void;
    protected handleCall(msg: CallMessage): void;
    protected handleResponse(msg: ResponseMessage): void;
    stop(): void;
    call(method: string, ...params: Vals): Promise<Vals>;
    protected raiseErr(err: Error): void;
}
export declare class ParentIPC extends IPCCommon {
    private readonly cmdPath;
    private readonly cmdArgs;
    private cmd;
    private readonly listener;
    constructor(cmdPath: string, cmdArgs: string[], ...localApis: object[]);
    start(): Promise<void>;
    private acceptConn;
    wait(): Promise<void>;
}
export declare class ChildIPC extends IPCCommon {
    constructor(...localApis: object[]);
    start(): Promise<void>;
    wait(): Promise<void>;
}
export {};
//# sourceMappingURL=lib.d.ts.map