"use client";

import { flexRender, getCoreRowModel, type ColumnDef, useReactTable } from "@tanstack/react-table";

export function DataTable<T>({ columns, data, label }: { columns: ColumnDef<T>[]; data: T[]; label: string }) {
  // TanStack Table intentionally exposes a mutable table instance; React Compiler skips this boundary.
  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({ columns, data, getCoreRowModel: getCoreRowModel() });
  return <div className="table-scroll" role="region" aria-label={label} tabIndex={0}><table className="data-table"><thead>{table.getHeaderGroups().map((group) => <tr key={group.id}>{group.headers.map((header) => <th key={header.id} scope="col">{header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}</th>)}</tr>)}</thead><tbody>{table.getRowModel().rows.map((row) => <tr key={row.id}>{row.getVisibleCells().map((cell) => <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>)}</tr>)}</tbody></table></div>;
}
