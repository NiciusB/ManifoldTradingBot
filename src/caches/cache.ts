// import required to polyfill localStorage in Deno Deploy
import "@sigma/deno-compile-extra/localStoragePolyfill";

// Generic cache implementation with TypeScript generics
export class Cache<T> {
  private namespace: string;
  private fetchFn: (key: string) => Promise<T>;
  private ttl: number;
  private minRefreshInterval: number;

  constructor(
    namespace: string,
    fetchFn: (key: string) => Promise<T>,
    ttl: number,
    minRefreshInterval: number,
  ) {
    this.namespace = namespace;
    this.fetchFn = fetchFn;
    this.ttl = ttl;
    this.minRefreshInterval = minRefreshInterval;
  }

  private _get(key: string): {
    value: T;
    savedAt: number;
  } | null {
    const val = localStorage.getItem(this.namespace + ":" + key);
    return val ? JSON.parse(val) : null;
  }

  private _set(key: string, value: T): void {
    localStorage.setItem(
      this.namespace + ":" + key,
      JSON.stringify({ value, savedAt: Date.now() }),
    );
  }

  async get(key: string): Promise<T> {
    const now = Date.now();
    const cached = this._get(key);

    if (cached && now < cached.savedAt + this.ttl) {
      if (now >= cached.savedAt + this.minRefreshInterval) {
        this.fetchFn(key).then((value) => {
          this._set(key, value);
        });
      }

      return cached.value;
    }

    const value = await this.fetchFn(key);
    this._set(key, value);
    return value;
  }

  async renew(key: string): Promise<void> {
    const value = await this.fetchFn(key);
    this._set(key, value);
  }
}
