export function CrosswordPreview({ value }: { value: unknown }) {
  return (
    <div className="p-4 bg-gray-50 border rounded-md">
      <h3 className="font-bold text-gray-700">CrosswordPreview</h3>
      <pre className="text-xs mt-2 overflow-auto max-h-40">{JSON.stringify(value, null, 2)}</pre>
    </div>
  );
}
