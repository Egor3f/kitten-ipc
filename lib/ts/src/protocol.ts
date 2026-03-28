export enum MsgType {
    Call = 1,
    Response = 2,
}

export type Vals = any[];

export interface CallMessage {
    type: MsgType.Call,
    id: number,
    method: string;
    args: Vals;
}

export interface ResponseMessage {
    type: MsgType.Response,
    id: number,
    result?: Vals;
    error?: string;
}

export type Message = CallMessage | ResponseMessage;

export interface CallResult {
    result: Vals;
    error: Error | null;
}
