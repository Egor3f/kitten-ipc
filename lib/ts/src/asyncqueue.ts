export class AsyncQueue<T> {
    private store: T[] = [];
    private collectors: ((val: T[]) => void)[] = [];

    put(val: T) {
        this.store.push(val);
        for(const collector of this.collectors) {
            collector(this.store);
        }
        this.collectors = [];
    }

    async collect(): Promise<T[]> {
        if(this.store.length > 0) {
            const store = this.store;
            this.store = [];
            return new Promise(resolve => resolve(store));
        } else {
            return new Promise(resolve => {
                this.collectors.push(resolve);
            })
        }
    }
}
