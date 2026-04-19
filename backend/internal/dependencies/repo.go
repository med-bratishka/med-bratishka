package dependencies

import (
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
)

func (d *Dependencies) TxRepo() transaction.Repository {
	if d.txRepo == nil {
		d.txRepo = transaction.NewTxRepository(d.pgClient)
	}
	return d.txRepo
}

func (d *Dependencies) UsersRepo() repository.UsersRepository {
	if d.usersRepo == nil {
		d.usersRepo = repository.NewUsersRepository(d.pgClient)
	}
	return d.usersRepo
}

func (d *Dependencies) SessionsRepo() repository.SessionsRepository {
	if d.sessionsRepo == nil {
		d.sessionsRepo = repository.NewSessionsRepository(d.pgClient)
	}
	return d.sessionsRepo
}

func (d *Dependencies) ClinicRepo() repository.ClinicRepository {
	if d.clinicRepo == nil {
		d.clinicRepo = repository.NewClinicRepository(d.pgClient)
	}
	return d.clinicRepo
}

func (d *Dependencies) DoctorRepo() repository.DoctorRepository {
	if d.doctorRepo == nil {
		d.doctorRepo = repository.NewDoctorRepository(d.pgClient)
	}
	return d.doctorRepo
}

func (d *Dependencies) PatientRepo() repository.PatientRepository {
	if d.patientRepo == nil {
		d.patientRepo = repository.NewPatientRepository(d.pgClient)
	}
	return d.patientRepo
}

func (d *Dependencies) ChatRepo() repository.ChatRepository {
	if d.chatRepo == nil {
		d.chatRepo = repository.NewChatRepository()
	}
	return d.chatRepo
}
