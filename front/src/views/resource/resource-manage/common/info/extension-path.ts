type FieldPath<T> = T extends object
  ? {
      [Key in keyof T & string]: NonNullable<T[Key]> extends (infer Item)[]
        ? Key | `${Key}[${number}]` | `${Key}[${number}].${FieldPath<NonNullable<Item>>}`
        : Key | `${Key}.${FieldPath<NonNullable<T[Key]>>}`;
    }[keyof T & string]
  : never;

export const extensionPath =
  <Resource extends { extension?: unknown }>() =>
  (path: 'extension' | `extension.${FieldPath<NonNullable<Resource['extension']>>}`) =>
    path;
