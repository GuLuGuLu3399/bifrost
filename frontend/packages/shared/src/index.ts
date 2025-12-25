export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue };

export type Branded<T, Brand extends string> = T & { __brand: Brand };
