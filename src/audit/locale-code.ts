export function isCanonicalAuditLocale(value: string | undefined): boolean {
  const code = value ?? "";
  return (
    code.length <= 64 && /^[A-Za-z0-9]{1,8}(-[A-Za-z0-9]{1,8})*$/.test(code)
  );
}
