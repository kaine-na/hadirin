<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { employeesApi } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import { isHR, isHRorManager } from '$lib/stores/auth.svelte';
	import Badge from '$lib/components/Badge.svelte';
	import Table from '$lib/components/Table.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import ConfirmModal from '$lib/components/ui/ConfirmModal.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import type { Column } from '$lib/components/types';
	import { formatDate, roleLabel } from '$lib/utils/format';
	import type { Employee, Role, CreateEmployeePayload, UpdateEmployeePayload } from '$lib/types';
	import {
		UserPlus,
		Search,
		Edit,
		UserX,
		Users,
		Filter,
		RefreshCw,
		ChevronDown
	} from 'lucide-svelte';

	let employees = $state<Employee[]>([]);
	let loading = $state(true);
	let search = $state('');
	let filterRole = $state('');
	let filterDept = $state('');

	// Pagination state — terhubung ke API params ?page=N&limit=N
	let currentPage = $state(1);
	const PAGE_SIZE = 20;
	let totalEmployees = $state(0);

	// Modal state
	let formOpen = $state(false);
	let editingId = $state<string | null>(null);
	let saving = $state(false);

	// Confirm deactivate
	let confirmOpen = $state(false);
	let confirmEmp = $state<Employee | null>(null);

	let fName = $state('');
	let fEmail = $state('');
	let fPassword = $state('');
	let fRole = $state<Role>('karyawan');
	let fDepartment = $state('');
	let fPosition = $state('');
	let fNIK = $state('');
	let fActive = $state(true);

	const roles: Role[] = ['karyawan', 'manager', 'hr_admin', 'super_admin'];

	const columns: Column<Employee>[] = [
		{ key: 'name', label: 'Nama' },
		{ key: 'email', label: 'Email' },
		{ key: 'role', label: 'Role' },
		{ key: 'department', label: 'Departemen' },
		{ key: 'position', label: 'Jabatan' },
		{ key: 'joined_at', label: 'Bergabung' },
		{ key: 'status', label: 'Status' },
		{ key: 'actions', label: '', align: 'right' }
	];

	async function load() {
		loading = true;
		try {
			const res = await employeesApi.list({
				search: search || undefined,
				role: filterRole || undefined,
				department: filterDept || undefined,
				page: currentPage,
				page_size: PAGE_SIZE
			});
			employees = res.items;
			totalEmployees = res.total ?? res.items.length;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal memuat karyawan');
		} finally {
			loading = false;
		}
	}

	function handlePageChange(page: number) {
		currentPage = page;
		load();
	}

	function handleFilterSearch() {
		// Reset ke halaman 1 saat filter berubah
		currentPage = 1;
		load();
	}

	onMount(() => {
		if (!isHRorManager()) {
			goto('/dashboard');
			return;
		}
		load();
	});

	function resetForm() {
		editingId = null;
		fName = '';
		fEmail = '';
		fPassword = '';
		fRole = 'karyawan';
		fDepartment = '';
		fPosition = '';
		fNIK = '';
		fActive = true;
	}

	function openCreate() {
		resetForm();
		formOpen = true;
	}

	function openEdit(emp: Employee) {
		editingId = emp.id;
		fName = emp.name;
		fEmail = emp.email;
		fPassword = '';
		fRole = emp.role;
		fDepartment = emp.department ?? '';
		fPosition = emp.position ?? '';
		fNIK = emp.nik ?? '';
		fActive = emp.is_active;
		formOpen = true;
	}

	async function save(e: SubmitEvent) {
		e.preventDefault();
		saving = true;
		try {
			if (editingId) {
				const payload: UpdateEmployeePayload = {
					name: fName,
					role: fRole,
					department: fDepartment,
					position: fPosition,
					nik: fNIK,
					is_active: fActive
				};
				await employeesApi.update(editingId, payload);
				toast.success('Data karyawan diperbarui');
			} else {
				if (!fPassword) {
					toast.error('Password wajib diisi untuk karyawan baru');
					saving = false;
					return;
				}
				const payload: CreateEmployeePayload = {
					name: fName,
					email: fEmail,
					password: fPassword,
					role: fRole,
					department: fDepartment,
					position: fPosition,
					nik: fNIK
				};
				await employeesApi.create(payload);
				toast.success('Karyawan baru berhasil ditambahkan');
			}
			formOpen = false;
			resetForm();
			await load();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Gagal menyimpan');
		} finally {
			saving = false;
		}
	}

	function confirmDeactivate(emp: Employee) {
		confirmEmp = emp;
		confirmOpen = true;
	}

	async function doDeactivate() {
		if (!confirmEmp) return;
		try {
			await employeesApi.remove(confirmEmp.id);
			toast.success('Karyawan berhasil dinonaktifkan');
			await load();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal');
		} finally {
			confirmOpen = false;
			confirmEmp = null;
		}
	}
</script>

<svelte:head>
	<title>Karyawan — Hadir</title>
</svelte:head>

<ConfirmModal
	open={confirmOpen}
	title="Nonaktifkan Karyawan"
	message={`Apakah Anda yakin ingin menonaktifkan karyawan "${confirmEmp?.name}"? Mereka tidak akan bisa login setelah ini.`}
	confirmLabel="Nonaktifkan"
	danger={true}
	onConfirm={doDeactivate}
	onCancel={() => {
		confirmOpen = false;
		confirmEmp = null;
	}}
/>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<PageHeader
			title="Karyawan"
			description="Kelola data seluruh karyawan perusahaan. Tambah, edit, dan pantau status karyawan."
		/>
		{#if isHR()}
			<button
				type="button"
				onclick={openCreate}
				class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-all duration-150 hover:bg-blue-700 hover:shadow-md active:scale-95"
			>
				<UserPlus size={16} strokeWidth={2} />
				Tambah Karyawan
			</button>
		{/if}
	</div>

	<!-- Filter Bar -->
	<div class="rounded-xl border border-slate-100 bg-white p-5 shadow-sm">
		<div class="grid gap-4 md:grid-cols-4">
			<!-- Search -->
			<div class="relative md:col-span-2">
				<Search
					size={16}
					strokeWidth={1.75}
					class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
				/>
				<input
					type="text"
					placeholder="Cari nama atau email karyawan..."
					class="w-full rounded-lg border border-slate-200 bg-white py-2 pl-9 pr-4 text-sm text-slate-900 placeholder-slate-400 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
					bind:value={search}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
			</div>

			<div class="relative">
				<select
					class="w-full appearance-none rounded-lg border border-slate-200 bg-white px-3 py-2 pr-8 text-sm text-slate-700 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
					bind:value={filterRole}
				>
					<option value="">Semua role</option>
					{#each roles as r}
						<option value={r}>{roleLabel(r)}</option>
					{/each}
				</select>
				<ChevronDown
					size={14}
					strokeWidth={2}
					class="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400"
				/>
			</div>

			<div class="flex items-center gap-2">
				<input
					type="text"
					placeholder="Filter departemen..."
					class="flex-1 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 shadow-sm transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
					bind:value={filterDept}
					onkeydown={(e) => e.key === 'Enter' && handleFilterSearch()}
				/>
				<button
					type="button"
					onclick={handleFilterSearch}
					disabled={loading}
					class="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition-all duration-150 hover:bg-slate-50 hover:shadow-md active:scale-95 disabled:opacity-50"
				>
					{#if loading}
						<RefreshCw size={14} strokeWidth={2} class="animate-spin" />
					{:else}
						<Filter size={14} strokeWidth={2} />
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Tabel Karyawan -->
	{#if !loading && employees.length === 0}
		<EmptyState
			icon={Users}
			title="Belum ada karyawan"
			description="Tambahkan karyawan pertama untuk memulai."
			ctaLabel={isHR() ? 'Tambah Karyawan' : undefined}
			onCta={isHR() ? openCreate : undefined}
		/>
	{:else}
		<div class="overflow-hidden rounded-xl border border-slate-100 bg-white shadow-sm">
			<Table {columns} rows={employees} loading={loading} emptyMessage="Belum ada karyawan">
				{#snippet row(e: Employee)}
					<tr class="group transition-colors hover:bg-slate-50">
						<td class="px-4 py-3">
							<div class="flex items-center gap-3">
								{#if e.photo_url}
									<img
										src={e.photo_url}
										alt={e.name}
										class="h-9 w-9 shrink-0 rounded-full object-cover ring-2 ring-white"
									/>
								{:else}
									<div
										class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-semibold text-blue-700 ring-2 ring-white"
									>
										{e.name.charAt(0).toUpperCase()}
									</div>
								{/if}
								<div>
									<p class="text-sm font-medium text-slate-900">{e.name}</p>
									{#if e.nik}
										<p class="text-xs text-slate-400">NIK: {e.nik}</p>
									{/if}
								</div>
							</div>
						</td>
						<td class="px-4 py-3 text-sm text-slate-600">{e.email}</td>
						<td class="px-4 py-3 text-sm">
							<Badge color="primary">{roleLabel(e.role)}</Badge>
						</td>
						<td class="px-4 py-3 text-sm text-slate-600">{e.department || '-'}</td>
						<td class="px-4 py-3 text-sm text-slate-600">{e.position || '-'}</td>
						<td class="px-4 py-3 text-sm text-slate-500">{formatDate(e.joined_at)}</td>
						<td class="px-4 py-3 text-sm">
							{#if e.is_active}
								<Badge color="green">Aktif</Badge>
							{:else}
								<Badge color="slate">Non-aktif</Badge>
							{/if}
						</td>
						<td class="px-4 py-3 text-right text-sm">
							{#if isHR()}
								<div
									class="flex justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100"
								>
									<button
										type="button"
										class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-blue-50 hover:text-blue-600"
										onclick={() => openEdit(e)}
										title="Edit karyawan"
									>
										<Edit size={14} strokeWidth={1.75} />
									</button>
									{#if e.is_active}
										<button
											type="button"
											class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-red-50 hover:text-red-600"
											onclick={() => confirmDeactivate(e)}
											title="Nonaktifkan"
										>
											<UserX size={14} strokeWidth={1.75} />
										</button>
									{/if}
								</div>
							{/if}
						</td>
					</tr>
				{/snippet}
			</Table>
			<!-- Pagination terhubung ke API params -->
			<Pagination
				page={currentPage}
				pageSize={PAGE_SIZE}
				total={totalEmployees}
				onPageChange={handlePageChange}
			/>
		</div>
	{/if}
</div>

<!-- Modal Form Karyawan -->
<Modal bind:open={formOpen} title={editingId ? 'Edit Karyawan' : 'Tambah Karyawan'} size="lg">
	<form id="emp-form" onsubmit={save} class="space-y-4">
		<div class="grid gap-4 md:grid-cols-2">
			<div>
				<label for="nm" class="label">Nama Lengkap <span class="text-red-500">*</span></label>
				<input id="nm" type="text" required class="input" bind:value={fName} />
			</div>
			<div>
				<label for="em" class="label">Email <span class="text-red-500">*</span></label>
				<input
					id="em"
					type="email"
					required
					class="input"
					bind:value={fEmail}
					disabled={!!editingId}
				/>
			</div>
			{#if !editingId}
				<div class="md:col-span-2">
					<label for="pw" class="label">Password <span class="text-red-500">*</span></label>
					<input
						id="pw"
						type="password"
						required
						class="input"
						minlength="6"
						bind:value={fPassword}
					/>
					<p class="mt-1 text-xs text-slate-500">Minimum 6 karakter.</p>
				</div>
			{/if}
			<div>
				<label for="rl" class="label">Role</label>
				<select id="rl" class="input" bind:value={fRole}>
					{#each roles as r}
						<option value={r}>{roleLabel(r)}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="nk" class="label">NIK</label>
				<input id="nk" type="text" class="input" bind:value={fNIK} />
			</div>
			<div>
				<label for="dp" class="label">Departemen</label>
				<input id="dp" type="text" class="input" bind:value={fDepartment} />
			</div>
			<div>
				<label for="ps" class="label">Jabatan</label>
				<input id="ps" type="text" class="input" bind:value={fPosition} />
			</div>
			{#if editingId}
				<div class="md:col-span-2">
					<label class="flex cursor-pointer items-center gap-2 text-sm text-slate-700">
						<input type="checkbox" class="rounded border-slate-300" bind:checked={fActive} />
						Karyawan aktif
					</label>
				</div>
			{/if}
		</div>
	</form>
	{#snippet footer()}
		<button
			type="button"
			onclick={() => (formOpen = false)}
			class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition-all hover:bg-slate-50 active:scale-95"
		>
			Batal
		</button>
		<button
			type="button"
			disabled={saving}
			onclick={() => {
				const form = document.getElementById('emp-form') as HTMLFormElement | null;
				form?.requestSubmit();
			}}
			class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-all hover:bg-blue-700 active:scale-95 disabled:opacity-50"
		>
			{#if saving}
				<RefreshCw size={14} strokeWidth={2} class="animate-spin" />
				Menyimpan...
			{:else}
				Simpan
			{/if}
		</button>
	{/snippet}
</Modal>
