# Generated by Django 5.0.6 on 2024-05-17 11:32

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('main', '0001_initial'),
    ]

    operations = [
        migrations.AddField(
            model_name='relevanceai',
            name='email_name',
            field=models.CharField(default='test', max_length=50),
        ),
        migrations.AddField(
            model_name='scrapin',
            name='email_name',
            field=models.CharField(default='test', max_length=50),
        ),
    ]
