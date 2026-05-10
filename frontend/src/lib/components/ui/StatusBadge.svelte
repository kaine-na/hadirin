<script lang="ts">
  interface Props {
    status: string;
  }

  let { status }: Props = $props();

  type BadgeStyle = { bg: string; text: string; label: string };

  const statusMap: Record<string, BadgeStyle> = {
    // Attendance
    hadir:        { bg: 'bg-green-100',  text: 'text-green-700',  label: 'Hadir' },
    present:      { bg: 'bg-green-100',  text: 'text-green-700',  label: 'Present' },
    terlambat:    { bg: 'bg-yellow-100', text: 'text-yellow-700', label: 'Terlambat' },
    late:         { bg: 'bg-yellow-100', text: 'text-yellow-700', label: 'Late' },
    alpha:        { bg: 'bg-red-100',    text: 'text-red-700',    label: 'Alpha' },
    absent:       { bg: 'bg-red-100',    text: 'text-red-700',    label: 'Absent' },
    izin:         { bg: 'bg-purple-100', text: 'text-purple-700', label: 'Izin' },
    permission:   { bg: 'bg-purple-100', text: 'text-purple-700', label: 'Permission' },
    sakit:        { bg: 'bg-blue-100',   text: 'text-blue-700',   label: 'Sakit' },
    sick:         { bg: 'bg-blue-100',   text: 'text-blue-700',   label: 'Sick' },
    // Request / approval
    menunggu:     { bg: 'bg-slate-100',  text: 'text-slate-600',  label: 'Menunggu' },
    pending:      { bg: 'bg-slate-100',  text: 'text-slate-600',  label: 'Pending' },
    disetujui:    { bg: 'bg-green-100',  text: 'text-green-700',  label: 'Disetujui' },
    approved:     { bg: 'bg-green-100',  text: 'text-green-700',  label: 'Approved' },
    ditolak:      { bg: 'bg-red-100',    text: 'text-red-700',    label: 'Ditolak' },
    rejected:     { bg: 'bg-red-100',    text: 'text-red-700',    label: 'Rejected' },
    perlu_revisi: { bg: 'bg-orange-100', text: 'text-orange-700', label: 'Perlu Revisi' },
    revision:     { bg: 'bg-orange-100', text: 'text-orange-700', label: 'Revision' },
    // Employee status
    aktif:        { bg: 'bg-green-100',  text: 'text-green-700',  label: 'Aktif' },
    active:       { bg: 'bg-green-100',  text: 'text-green-700',  label: 'Active' },
    nonaktif:     { bg: 'bg-red-100',    text: 'text-red-700',    label: 'Nonaktif' },
    inactive:     { bg: 'bg-red-100',    text: 'text-red-700',    label: 'Inactive' },
  };

  const badge = $derived(
    statusMap[status.toLowerCase()] ?? {
      bg: 'bg-slate-100',
      text: 'text-slate-600',
      label: status,
    }
  );
</script>

<span
  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {badge.bg} {badge.text}"
>
  {badge.label}
</span>
