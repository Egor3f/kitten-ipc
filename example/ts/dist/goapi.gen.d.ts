import { ParentIPC, ChildIPC } from "../../lib/ts/lib";
export default class IpcApi {
    private ipc;
    constructor(ipc: ParentIPC | ChildIPC);
    Div(a: number, b: number): Promise<number>;
}
//# sourceMappingURL=goapi.gen.d.ts.map