package models

import (
	"io"
	"strconv"
)

type User struct {
	ID     int
	Name   string
	Email  string
	Role   string
	Status string
	Avatar string
}

var Users = []User{
	{1, "Ana Silva", "ana@exemplo.com", "Admin", "active", "AS"},
	{2, "Bruno Costa", "bruno@exemplo.com", "Editor", "active", "BC"},
	{3, "Carla Mendes", "carla@exemplo.com", "Viewer", "inactive", "CM"},
	{4, "Diego Souza", "diego@exemplo.com", "Editor", "active", "DS"},
	{5, "Elisa Prado", "elisa@exemplo.com", "Viewer", "active", "EP"},
}

func Contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func Initials(name string) string {
	if name == "" {
		return "?"
	}
	parts := []rune(name)
	if len(parts) >= 2 {
		return string(parts[0]) + string(parts[len(parts)-1])
	}
	return string(parts[0])
}

func FindUserByID(id int) (User, bool) {
	for _, u := range Users {
		if u.ID == id {
			return u, true
		}
	}
	return User{}, false
}

func UpdateUser(id int, name, email, role, status string) bool {
	for i := range Users {
		if Users[i].ID == id {
			Users[i].Name = name
			Users[i].Email = email
			Users[i].Role = role
			Users[i].Status = status
			Users[i].Avatar = Initials(name)
			return true
		}
	}
	return false
}

func RenderUserTable(w io.Writer, users []User) {
	io.WriteString(w, `<table class="min-w-full divide-y divide-gray-200"><thead class="bg-gray-50"><tr><th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Usuario</th><th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role</th><th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th><th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Acoes</th></tr></thead><tbody class="bg-white divide-y divide-gray-200">`)
	for _, u := range users {
		RenderUserRow(w, u)
	}
	io.WriteString(w, `</tbody></table>`)
}

func RenderUserRow(w io.Writer, u User) {
	statusBadge := `<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800">Inativo</span>`
	if u.Status == "active" {
		statusBadge = `<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">Ativo</span>`
	}
	idStr := strconv.Itoa(u.ID)
	io.WriteString(w, `<tr id="user-row-`+idStr+`"><td class="px-6 py-4 whitespace-nowrap"><div class="flex items-center"><div class="flex-shrink-0 h-10 w-10"><div class="h-10 w-10 rounded-full bg-blue-600 flex items-center justify-center text-white font-medium text-sm">`+u.Avatar+`</div></div><div class="ml-4"><div class="text-sm font-medium text-gray-900">`+u.Name+`</div><div class="text-sm text-gray-500">`+u.Email+`</div></div></div></td><td class="px-6 py-4 whitespace-nowrap"><span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">`+u.Role+`</span></td><td class="px-6 py-4 whitespace-nowrap">`+statusBadge+`</td><td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500"><button class="text-blue-600 hover:text-blue-900 mr-3" hx-get="/user/edit?id=`+idStr+`" hx-target="#modal-container" hx-swap="innerHTML">Editar</button><button class="text-red-600 hover:text-red-900" hx-delete="/api/dashboard?id=`+idStr+`" hx-target="closest tr" hx-swap="delete" hx-confirm="Tem certeza que deseja excluir este usuario?">Excluir</button></td></tr>`)
}

func RenderUserEditForm(w io.Writer, u User) {
	selected := func(role, target string) string {
		if role == target {
			return " selected"
		}
		return ""
	}
	idStr := strconv.Itoa(u.ID)
	io.WriteString(w, `<div class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50" hx-target="this" hx-swap="outerHTML"><div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4"><div class="px-6 py-4 border-b border-gray-200"><h3 class="text-lg font-medium text-gray-900">Editar Usuario</h3></div><form class="px-6 py-4 space-y-4" hx-put="/api/dashboard" hx-target="#user-row-`+idStr+`" hx-swap="outerHTML" hx-on::after-request="if (document.querySelector('.fixed')) document.querySelector('.fixed').remove()"><input type="hidden" name="id" value="`+idStr+`"/><div><label class="block text-sm font-medium text-gray-700">Nome</label><input type="text" name="name" value="`+u.Name+`" required class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"/></div><div><label class="block text-sm font-medium text-gray-700">Email</label><input type="email" name="email" value="`+u.Email+`" required class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"/></div><div><label class="block text-sm font-medium text-gray-700">Role</label><select name="role" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"><option value="Viewer"`+selected(u.Role, "Viewer")+`>Viewer</option><option value="Editor"`+selected(u.Role, "Editor")+`>Editor</option><option value="Admin"`+selected(u.Role, "Admin")+`>Admin</option></select></div><div><label class="block text-sm font-medium text-gray-700">Status</label><select name="status" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"><option value="active"`+selected(u.Status, "active")+`>Ativo</option><option value="inactive"`+selected(u.Status, "inactive")+`>Inativo</option></select></div><div class="flex justify-end gap-3 pt-4"><button type="button" class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50" onclick="this.closest('.fixed').remove()">Cancelar</button><button type="submit" class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">Salvar</button></div></form></div></div>`)
}

func RenderUserForm(w io.Writer) {
	io.WriteString(w, `<div class="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50" hx-target="this" hx-swap="outerHTML"><div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4"><div class="px-6 py-4 border-b border-gray-200"><h3 class="text-lg font-medium text-gray-900">Novo Usuario</h3></div><form class="px-6 py-4 space-y-4" hx-post="/api/dashboard" hx-target="#users-table" hx-swap="innerHTML" hx-on::after-request="this.closest('.fixed').remove()"><div><label class="block text-sm font-medium text-gray-700">Nome</label><input type="text" name="name" required class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"/></div><div><label class="block text-sm font-medium text-gray-700">Email</label><input type="email" name="email" required class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"/></div><div><label class="block text-sm font-medium text-gray-700">Role</label><select name="role" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"><option value="Viewer">Viewer</option><option value="Editor">Editor</option><option value="Admin">Admin</option></select></div><div class="flex justify-end gap-3 pt-4"><button type="button" class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50" onclick="this.closest('.fixed').remove()">Cancelar</button><button type="submit" class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">Salvar</button></div></form></div></div>`)
}
