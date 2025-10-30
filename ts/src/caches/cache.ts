// Generic cache implementation with TypeScript generics
export class Cache<T> {
  private cache: Map<string, { value: T; savedAt: number }> = new Map();
  private fetchFn: (key: string) => Promise<T>;
  private ttl: number;
  private minRefreshInterval: number;

  constructor(
    fetchFn: (key: string) => Promise<T>,
    ttl: number,
    minRefreshInterval: number,
  ) {
    this.fetchFn = fetchFn;
    this.ttl = ttl;
    this.minRefreshInterval = minRefreshInterval;
  }

  async get(key: string): Promise<T> {
    const now = Date.now();
    const cached = this.cache.get(key);

    if (cached && now < cached.savedAt + this.ttl) {
      if (now >= cached.savedAt + this.minRefreshInterval) {
        this.fetchFn(key).then((value) => {
          this.cache.set(key, {
            value,
            savedAt: Date.now(),
          });
        });
      }

      return cached.value;
    }

    const value = await this.fetchFn(key);
    this.cache.set(key, {
      value,
      savedAt: now,
    });
    return value;
  }

  async renew(key: string): Promise<void> {
    const value = await this.fetchFn(key);
    this.cache.set(key, {
      value,
      savedAt: Date.now(),
    });
  }
}
